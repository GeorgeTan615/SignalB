package marketprice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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
	query := `select class 
					from ticker
					where symbol = ?`
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	var class string

	err := database.MySqlDB.QueryRowContext(ctx, query, ticker).Scan(&class)
	if err != nil {
		return "", err
	}

	return class, nil
}

func refreshData(c context.Context, ticker, timeframe string, data []*TickerData) error {
	// Based on how much data we adding, we will remove x amount of data
	count := len(data)
	table := "price_" + strings.ToLower(timeframe)
	delQuery := fmt.Sprintf(`delete 
						from %s
						where ticker_symbol = ?
						order by time 
						limit ?`, table)

	delCtx, cancel := context.WithTimeout(c, 10*time.Second)
	defer cancel()

	_, err := database.MySqlDB.ExecContext(delCtx, delQuery, ticker, count)
	if err != nil {
		return err
	}

	// Batch our inserts together to make our write more efficient
	var builder strings.Builder
	insQuery := `insert into %s (ticker_symbol,time,price) values ('%s','%s',%.2f);`
	for i := len(data) - 1; i > -1; i-- {
		currData := data[i]

		timeString := currData.Time.Format("2006-01-02 15:04:05")
		nxtQuery := fmt.Sprintf(insQuery, table, ticker, timeString, currData.Price)
		builder.WriteString(nxtQuery)
	}

	finalQuery := builder.String()
	tx, err := database.MySqlDB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(finalQuery)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// TODO refactor
func refreshPriceByTimeframe(c context.Context, timeframe string) ([]*RefreshPriceResp, error) {
	// Get all ticker along with class
	tickers, err := getTickersByTimeframe(c, timeframe)
	if err != nil {
		return nil, err
	}

	// For each ticker, refresh the data
	var results []*RefreshPriceResp
	chRes := make(chan *RefreshPriceResp, len(tickers))
	chErr := make(chan error)
	var wg sync.WaitGroup

	for _, ticker := range tickers {
		wg.Add(1)
		go func(c context.Context, ticker *database.Ticker, timeframe string, chRes chan<- *RefreshPriceResp, chErr chan<- error) {
			ctx, cancel := context.WithTimeout(c, 5*time.Second)
			defer cancel()

			result, err := refreshPriceByTickerClassTimeframe(ctx, ticker.Symbol, ticker.Class, timeframe)

			if err != nil {
				chErr <- err
				log.Printf("Error refreshing price for %s %s %s", ticker.Symbol, timeframe, err)
			} else {
				chRes <- result
				log.Printf("Finished refreshing price for %s %s", ticker.Symbol, timeframe)
			}
		}(c, ticker, timeframe, chRes, chErr)
	}

	timeTicker := time.NewTicker(10 * time.Second)
	defer timeTicker.Stop()
	for {
		select {
		case res := <-chRes:
			results = append(results, res)
			if len(results) == len(tickers) {
				return results, nil
			}

		case err := <-chErr:
			return nil, err

		case <-timeTicker.C:
			return nil, errors.New("exceeded 1 minute for refreshing data")
		}
	}
}

func getTickersByTimeframe(c context.Context, timeframe string) ([]*database.Ticker, error) {
	query := `select distinct t.symbol, t.class
					from ticker t join binding b on t.symbol = b.ticker_symbol
					where b.timeframe = ?`

	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	res, err := database.MySqlDB.QueryContext(ctx, query, timeframe)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var tickers []*database.Ticker
	for res.Next() {
		var ticker database.Ticker

		err = res.Scan(&ticker.Symbol, &ticker.Class)
		if err != nil {
			return nil, err
		}

		tickers = append(tickers, &ticker)
	}

	return tickers, nil
}
