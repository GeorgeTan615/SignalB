package strategy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/signalb/internal/database"
)

type Resp struct {
	Strategy          Strategy `json:"strategy"`
	IsFulfilled       bool     `json:"isFulfilled"`
	EvaluationMessage string   `json:"evaluationMessage"`
}

type tickerStrategiesResult struct {
	tickerSymbol     string
	strategiesResult []*Resp
}

func evaluateTickersStrategiesByTimeframe(c context.Context, timeframe string) (map[string][]*Resp, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	tickersStrategiesMap, err := getTickersAndStrategyByTimeframe(ctx, timeframe)
	if err != nil {
		return nil, err
	}

	var (
		chRes      = make(chan tickerStrategiesResult, len(tickersStrategiesMap))
		chErr      = make(chan error, len(tickersStrategiesMap))
		result     = make(map[string][]*Resp)
		wgEvaluate sync.WaitGroup
		wgCollect  sync.WaitGroup
	)

	wgCollect.Add(2)
	// collect result
	go func() {
		defer wgCollect.Done()
		for res := range chRes {
			result[res.tickerSymbol] = res.strategiesResult
		}
	}()

	// collect error
	go func() {
		defer wgCollect.Done()
		for newErr := range chErr {
			err = newErr
		}
	}()

	for tickerSymbol, strategies := range tickersStrategiesMap {
		wgEvaluate.Add(1)
		go func(c context.Context, tickerSymbol, timeframe string, strategies []Strategy) {
			defer wgEvaluate.Done()

			ctx, cancel := context.WithTimeout(c, 4*time.Second)
			defer cancel()

			data, err := getTickerDataByTimeframe(ctx, tickerSymbol, timeframe)
			if err != nil {
				chErr <- err
				return
			}

			err = evaluateStrategiesForTicker(ctx, tickerSymbol, strategies, data, chRes)
			if err != nil {
				chErr <- err
			}
		}(c, tickerSymbol, timeframe, strategies)
	}

	wgEvaluate.Wait()
	close(chRes)
	close(chErr)
	wgCollect.Wait()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func getTickersAndStrategyByTimeframe(c context.Context, timeframe string) (map[string][]Strategy, error) {
	query := `select ticker_symbol, strategy
					from binding
					where timeframe = ?`

	res, err := database.Client.DB.QueryContext(c, query, timeframe)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	tickerToStrategiesMap := make(map[string][]Strategy)

	for res.Next() {
		var (
			tickerSymbol string
			strategyStr  string
		)

		err := res.Scan(&tickerSymbol, &strategyStr)
		if err != nil {
			return nil, err
		}

		strategy, err := strategyManager.GetStrategyByName(strategyStr)
		if err != nil {
			return nil, err
		}

		tickerToStrategiesMap[tickerSymbol] = append(tickerToStrategiesMap[tickerSymbol], strategy)
	}

	return tickerToStrategiesMap, nil
}

func evaluateStrategiesForTicker(
	c context.Context,
	tickerSymbol string,
	strategies []Strategy,
	data []float64,
	chRes chan<- tickerStrategiesResult,
) error {
	select {
	case <-c.Done():
		return c.Err()
	default:
	}

	var (
		chStrategyResp = make(chan *Resp)
		strategyResps  []*Resp
		wgEvaluate     sync.WaitGroup
		wgCollect      sync.WaitGroup
	)

	wgCollect.Add(1)
	go func() {
		defer wgCollect.Done()
		for resp := range chStrategyResp {
			strategyResps = append(strategyResps, resp)
		}
	}()

	for _, strategy := range strategies {
		wgEvaluate.Add(1)
		evaluateStrategy(c, data, strategy, chStrategyResp, &wgEvaluate)
	}

	wgEvaluate.Wait()
	close(chStrategyResp)
	wgCollect.Wait()

	if len(strategyResps) != len(strategies) {
		return errors.New("something went wrong")
	}

	chRes <- tickerStrategiesResult{
		tickerSymbol:     tickerSymbol,
		strategiesResult: strategyResps,
	}

	select {
	case <-c.Done():
		return c.Err()
	default:
		return nil
	}
}

func getTickerDataByTimeframe(c context.Context, tickerSymbol, timeframe string) ([]float64, error) {
	query := fmt.Sprintf(`select price
							from price_%s
							where ticker_symbol = ?
							order by time`, strings.ToLower(timeframe))

	rows, err := database.Client.DB.QueryContext(c, query, tickerSymbol)
	if err != nil {
		return nil, err
	}

	var prices []float64

	for rows.Next() {
		var price float64

		err := rows.Scan(&price)
		if err != nil {
			return nil, err
		}

		prices = append(prices, price)
	}

	return prices, nil
}

func evaluateStrategy(
	c context.Context,
	data []float64,
	strategy Strategy,
	chStrategyRes chan<- *Resp,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	select {
	case <-c.Done():
		return
	default:
	}

	result := strategy.Evaluate(data)

	strategyRes := &Resp{
		Strategy:          strategy,
		IsFulfilled:       result.IsFulfilled,
		EvaluationMessage: result.EvaluationMessage,
	}

	chStrategyRes <- strategyRes
}
