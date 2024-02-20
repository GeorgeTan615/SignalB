package marketprice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/signalb/internal/ticker"
	"github.com/signalb/internal/timeframe"
)

const (
	rapidApiIntradayMaximumLength = 250 // Intraday max is 1000
)

type RapidApiCredentials struct {
	baseUrl string
	key     string
	host    string
}

type StockDataFetcher struct {
	credentials      *RapidApiCredentials
	timeframeMapping map[string]string
}

func NewStockDataFetcher(credentials *RapidApiCredentials) *StockDataFetcher {
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

func (stockDF *StockDataFetcher) Fetch(timeframe, tickerSymbol string, length int) ([]*TickerData, error) {
	if length > RefreshAllDataLength {
		return nil, fmt.Errorf("maximum length is %d", RefreshAllDataLength)
	}

	var (
		results       []*TickerData
		err           error
		fetchStrategy func(credentials *RapidApiCredentials, timeframeVal, tickerSymbol string, length int) ([]*TickerData, error)
	)

	timeframeVal := stockDF.timeframeMapping[timeframe]
	if timeframeVal != "intraday" {
		fetchStrategy = handleNonIntradayDataFetching
	} else {
		fetchStrategy = handleIntradayDataFetching
	}

	results, err = fetchStrategy(stockDF.credentials, timeframeVal, tickerSymbol, length)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func handleIntradayDataFetching(credentials *RapidApiCredentials, timeframeVal, tickerSymbol string, length int) ([]*TickerData, error) {
	location, err := time.LoadLocation("America/New_York")

	if err != nil {
		return nil, err
	}

	if length > rapidApiIntradayMaximumLength {
		length = rapidApiIntradayMaximumLength
	}

	adjustedLength := 4 * length
	url := fmt.Sprintf("%s/%s?symbol=%s&interval=60min&maxreturn=%d", credentials.baseUrl, timeframeVal, tickerSymbol, adjustedLength)

	resp, err := makeRapidAPIHistoricalDataCall(url, credentials.key, credentials.host)

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

func handleNonIntradayDataFetching(credentials *RapidApiCredentials, timeframeVal, tickerSymbol string, length int) ([]*TickerData, error) {
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

	url := fmt.Sprintf("%s/%s?symbol=%s&dateStart=%s&dateEnd=%s", credentials.baseUrl, timeframeVal, tickerSymbol, dateStart, dateEnd)

	resp, err := makeRapidAPIHistoricalDataCall(url, credentials.key, credentials.host)

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

func makeRapidAPIHistoricalDataCall(url, key, host string) (*RapidApiDataResp, error) {
	req, err := http.NewRequest("GET", url, nil)

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

	var resp RapidApiDataResp

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
	if secondLastResult.Close == lastResult.Close {
		return len(results) - 2, weeklyJumpInterval
	} else {
		return len(results) - 1, weeklyJumpInterval
	}
}

func getDateStrToTime(layout string, loc *time.Location, dateStr string) (time.Time, error) {
	return time.ParseInLocation(layout, dateStr, loc)
}
