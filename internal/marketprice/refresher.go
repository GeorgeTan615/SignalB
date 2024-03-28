package marketprice

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/signalb/internal/database"
)

const (
	RefreshAllDataLength   = 300
	UpdateLatestDataLength = 1
)

func refreshPriceByTickerTimeframe(c context.Context, ticker, timeframe string) (*RefreshPriceResp, error) {
	class, err := getTickerClass(c, ticker)
	if err != nil {
		return nil, err
	}

	return refreshPriceByTickerClassTimeframe(c, ticker, class, timeframe)
}

func refreshPriceByTickerClassTimeframe(
	ctx context.Context,
	ticker, class, timeframe string,
) (*RefreshPriceResp, error) {
	// Just get all data since its inexpensive, dont have to deal with stale data
	length := RefreshAllDataLength

	// Get the fetcher we need based on class as we have different ways of fetching data
	fetcher, ok := fetcherManager.getFetcherByTickerClass(class)

	if !ok {
		return nil, errors.New("can't get data fetcher")
	}

	res, err := fetcher.Fetch(ctx, timeframe, ticker, length)
	if err != nil {
		return nil, err
	}

	err = refreshData(ctx, ticker, timeframe, res)
	if err != nil {
		return nil, err
	}

	return &RefreshPriceResp{
		Ticker:          ticker,
		Class:           class,
		Timeframe:       timeframe,
		RefreshedPrices: res,
	}, nil
}

func getTickerClass(c context.Context, ticker string) (string, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	return database.Client.GetTickerClassBySymbol(ctx, ticker)
}

func refreshData(c context.Context, tickerSymbol, timeframe string, data []*TickerData) error {
	// Based on how much data we adding, we will remove x amount of data
	ctx, cancel := context.WithTimeout(c, 20*time.Second)
	defer cancel()

	err := database.Client.DeletePriceData(ctx, tickerSymbol, timeframe, len(data))
	if err != nil {
		return err
	}

	var priceData []database.PriceData
	for _, d := range data {
		priceData = append(priceData, database.PriceData{
			TickerSymbol: tickerSymbol,
			Time:         d.Time,
			Price:        d.Price,
		})
	}

	return database.Client.InsertPriceData(ctx, timeframe, priceData)
}

func refreshPriceByTimeframe(c context.Context, timeframe string) ([]*RefreshPriceResp, error) {
	// get all ticker along with class
	tickers, err := getTickersByTimeframe(c, timeframe)
	if err != nil {
		return nil, err
	}

	var (
		results   []*RefreshPriceResp
		chRes     = make(chan *RefreshPriceResp, len(tickers))
		chErr     = make(chan error, len(tickers))
		wgRefresh sync.WaitGroup
		wgCollect sync.WaitGroup
	)

	wgCollect.Add(2)
	// collect results
	go func() {
		defer wgCollect.Done()
		for res := range chRes {
			results = append(results, res)
		}
	}()

	// collect errors
	go func() {
		defer wgCollect.Done()
		for newErr := range chErr {
			err = newErr
		}
	}()

	for _, ticker := range tickers {
		wgRefresh.Add(1)
		go func(ticker *database.Ticker, timeframe string, chRes chan<- *RefreshPriceResp, chErr chan<- error) {
			defer wgRefresh.Done()

			ctx, cancel := context.WithTimeout(c, 10*time.Second)
			defer cancel()

			result, err := refreshPriceByTickerClassTimeframe(ctx, ticker.Symbol, ticker.Class, timeframe)

			if err != nil {
				chErr <- err
				log.Printf("Error refreshing price for %s %s %s", ticker.Symbol, timeframe, err)
			} else {
				chRes <- result
				log.Printf("Finished refreshing price for %s %s", ticker.Symbol, timeframe)
			}
		}(ticker, timeframe, chRes, chErr)
	}

	wgRefresh.Wait()
	close(chRes)
	close(chErr)
	wgCollect.Wait()

	if err != nil {
		return nil, err
	}

	return results, err
}

func getTickersByTimeframe(c context.Context, timeframe string) ([]*database.Ticker, error) {
	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	return database.Client.GetTickersByTimeframe(ctx, timeframe)
}
