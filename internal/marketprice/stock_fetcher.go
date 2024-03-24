package marketprice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/signalb/internal/ticker"
	"github.com/signalb/internal/timeframe"
)

const (
	rapidAPIIntradayMaximumLength = 250 // Intraday max is 1000
)

type RapidAPICredentials struct {
	baseURL string
	key     string
	host    string
}

type StockDataFetcher struct {
	credentials      *RapidAPICredentials
	timeframeMapping map[string]string
}

func NewStockDataFetcher(credentials *RapidAPICredentials) *StockDataFetcher {
	timeframeMapping := map[string]string{
		timeframe.Day1:  "daily",
		timeframe.Week1: "weekly",
		timeframe.Hour4: "intraday",
	}

	return &StockDataFetcher{
		credentials:      credentials,
		timeframeMapping: timeframeMapping,
	}
}

func (stockDF *StockDataFetcher) FetchClass() string {
	return ticker.StockClass
}

func (stockDF *StockDataFetcher) Fetch(
	ctx context.Context,
	timeframe, tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	if length > RefreshAllDataLength {
		return nil, fmt.Errorf("maximum length is %d", RefreshAllDataLength)
	}

	var (
		results       []*TickerData
		err           error
		fetchStrategy func(ctx context.Context, credentials *RapidAPICredentials, timeframeVal, tickerSymbol string, length int) ([]*TickerData, error)
	)

	timeframeVal := stockDF.timeframeMapping[timeframe]
	if timeframeVal != "intraday" {
		fetchStrategy = handleNonIntradayDataFetching
	} else {
		fetchStrategy = handleIntradayDataFetching
	}

	results, err = fetchStrategy(ctx, stockDF.credentials, timeframeVal, tickerSymbol, length)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func handleIntradayDataFetching(
	ctx context.Context,
	credentials *RapidAPICredentials,
	timeframeVal, tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}

	if length > rapidAPIIntradayMaximumLength {
		length = rapidAPIIntradayMaximumLength
	}

	adjustedLength := 4 * length
	url := fmt.Sprintf("%s/%s?symbol=%s&interval=60min&maxreturn=%d", credentials.baseURL, timeframeVal, tickerSymbol, adjustedLength)

	resp, err := makeRapidAPIHistoricalDataCall(ctx, url, credentials.key, credentials.host)
	if err != nil {
		return nil, err
	}

	var results []*TickerData
	for i := len(resp.Results) - 1; len(results) < length; i -= 4 {
		currRes := resp.Results[i]
		parsedTime, err := getDateStrToTime("2006-01-02 15:00", location, currRes.Date)
		if err != nil {
			return nil, err
		}

		results = append(results, NewTickerData(parsedTime, currRes.Close))
	}

	return results, nil
}

func handleNonIntradayDataFetching(
	ctx context.Context,
	credentials *RapidAPICredentials,
	timeframeVal, tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}

	today := time.Now().In(location)
	lengthMultiplier := 2 // This is to ensure we always get enough data points
	var yesterday time.Time

	if timeframeVal == "daily" {
		yesterday = today.AddDate(0, 0, -length*lengthMultiplier)
	} else {
		yesterday = today.AddDate(0, 0, -length*lengthMultiplier*7)
	}

	dateEnd := fmt.Sprintf("%v-%v-%v", today.Year(), int(today.Month()), today.Day())
	dateStart := fmt.Sprintf("%v-%v-%v", yesterday.Year(), int(yesterday.Month()), yesterday.Day())

	url := fmt.Sprintf("%s/%s?symbol=%s&dateStart=%s&dateEnd=%s", credentials.baseURL, timeframeVal, tickerSymbol, dateStart, dateEnd)

	resp, err := makeRapidAPIHistoricalDataCall(ctx, url, credentials.key, credentials.host)
	if err != nil {
		return nil, err
	}

	var results []*TickerData
	idx, jumpInterval := getStartIndexAndJumpInterval(timeframeVal, resp.Results)

	for i := idx; len(results) < length; i -= jumpInterval {
		currRes := resp.Results[i]
		parsedTime, err := getDateStrToTime("2006-01-02", location, currRes.Date)
		if err != nil {
			return nil, err
		}

		results = append(results, NewTickerData(parsedTime, currRes.Close))
	}

	return results, nil
}

func makeRapidAPIHistoricalDataCall(ctx context.Context, url, key, host string) (*RapidAPIDataResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", key)
	req.Header.Add("X-RapidAPI-Host", host)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resp RapidAPIDataResp

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func getStartIndexAndJumpInterval(timeframeVal string, results []Result) (int, int) {
	if timeframeVal == "daily" {
		return len(results) - 1, 1
	}

	// For weekly data, the data is not provided in exact weeks, thus we need to do some skipping
	weeklyJumpInterval := 2
	if len(results) < 2 {
		return len(results) - 1, weeklyJumpInterval
	}

	lastResult := results[len(results)-1]
	secondLastResult := results[len(results)-2]

	// For weekly data, we need to see which data to start jumping from
	var startIdx int
	if secondLastResult.Close == lastResult.Close {
		startIdx = len(results) - 2
	} else {
		startIdx = len(results) - 1
	}

	return startIdx, weeklyJumpInterval
}

func getDateStrToTime(layout string, loc *time.Location, dateStr string) (time.Time, error) {
	return time.ParseInLocation(layout, dateStr, loc)
}
