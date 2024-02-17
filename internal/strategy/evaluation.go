package strategy

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/signalb/internal/database"
)

type StrategyResp struct {
	Strategy          Strategy `json:"strategy"`
	IsFulfilled       bool     `json:"isFulfilled"`
	EvaluationMessage string   `json:"evaluationMessage"`
}

type tickerStrategiesResult struct {
	tickerSymbol     string
	strategiesResult []*StrategyResp
}

func evaluateTickersStrategiesByTimeframe(timeframe string) (map[string][]*StrategyResp, error) {
	tickersStrategiesMap, err := getTickersAndStrategyByTimeframe(timeframe)

	if err != nil {
		return nil, err
	}

	chRes := make(chan tickerStrategiesResult)
	chErr := make(chan error)
	for tickerSymbol, strategies := range tickersStrategiesMap {
		go evaluateTickerWithStrategiesByTimeframe(tickerSymbol, strategies, timeframe, chRes, chErr)
	}

	result := make(map[string][]*StrategyResp)
	duration := 1 * time.Minute
	timer := time.NewTicker(duration)

	for {
		select {
		case res := <-chRes:
			result[res.tickerSymbol] = res.strategiesResult
			if len(tickersStrategiesMap) == len(result) {
				return result, nil
			}

		case err := <-chErr:
			return nil, err

		case <-timer.C:
			return nil, fmt.Errorf("evaluate strategies took longer than %s", duration)
		}
	}
}

func getTickersAndStrategyByTimeframe(timeframe string) (map[string][]Strategy, error) {
	query := `select ticker_symbol, strategy
					from binding
					where timeframe = ?`

	res, err := database.MySqlDB.Query(query, timeframe)

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

func evaluateTickerWithStrategiesByTimeframe(
	tickerSymbol string,
	strategies []Strategy,
	timeframe string,
	chRes chan<- tickerStrategiesResult,
	chErr chan<- error) {
	data, err := getTickerDataByTimeframe(tickerSymbol, timeframe)

	if err != nil {
		chErr <- err
		return
	}

	chStrategyResp := make(chan *StrategyResp)
	var strategyResps []*StrategyResp

	for _, strategy := range strategies {
		go evaluateDataWithStrategy(data, strategy, chStrategyResp)
	}

	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case resp := <-chStrategyResp:
			strategyResps = append(strategyResps, resp)
			if len(strategyResps) == len(strategies) {
				chRes <- tickerStrategiesResult{
					tickerSymbol:     tickerSymbol,
					strategiesResult: strategyResps,
				}
				return
			}
		case <-timer.C:
			chErr <- errors.New("took too long to evaluate strategies")
			return
		}
	}
}

func getTickerDataByTimeframe(tickerSymbol, timeframe string) ([]float64, error) {
	query := fmt.Sprintf(
		`select price
			from price_%s
			where ticker_symbol = ?
			order by time`, strings.ToLower(timeframe))

	rows, err := database.MySqlDB.Query(query, tickerSymbol)

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

func evaluateDataWithStrategy(
	data []float64,
	strategy Strategy,
	chStrategyRes chan<- *StrategyResp) {

	result := strategy.Evaluate(data)

	strategyRes := &StrategyResp{
		Strategy:          strategy,
		IsFulfilled:       result.IsFulfilled,
		EvaluationMessage: result.EvaluationMessage,
	}

	chStrategyRes <- strategyRes
}
