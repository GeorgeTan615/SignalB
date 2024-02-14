package marketprice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/signalb/internal/ticker"
	timeframePkg "github.com/signalb/internal/timeframe"
)

type TokenInsightCredentials struct {
	baseUrl string
	key     string
}

type CryptoDataFetcher struct {
	credentials *TokenInsightCredentials
}

func NewCryptoDataFetcher(credentials *TokenInsightCredentials) *CryptoDataFetcher {
	return &CryptoDataFetcher{
		credentials: credentials,
	}
}

func (cryptoDF *CryptoDataFetcher) FetchClass() string {
	return ticker.CryptoClass
}

func (cryptoDF *CryptoDataFetcher) Fetch(timeframe, tickerSymbol string, length int) ([]*TickerData, error) {
	if length > 200 {
		return nil, errors.New("maximum length is 200")
	}

	var (
		fetchStrategy func(credentials *TokenInsightCredentials, tickerSymbol string, length int) ([]*TickerData, error)
		err           error
	)

	switch {
	case timeframe == timeframePkg.Day1:
		fetchStrategy = handleDay1DataFetching
	case timeframe == timeframePkg.Hour4:
		fetchStrategy = handleHour4DataFetching
	case timeframe == timeframePkg.Week1:
		fetchStrategy = handleWeek1DataFetching
	default:
		err = errors.New("error getting timeframe mapping")
	}

	if err != nil {
		return nil, err
	}

	return fetchStrategy(cryptoDF.credentials, tickerSymbol, length)
}

func handleDay1DataFetching(credentials *TokenInsightCredentials, tickerSymbol string, length int) ([]*TickerData, error) {
	interval := "day"
	url := fmt.Sprintf("%s/%s?interval=%s&length=%d", credentials.baseUrl, strings.ToLower(tickerSymbol), interval, length)

	resp, err := makeTokenInsightHistoricalDataCall(url, credentials.key)

	if err != nil {
		return nil, err
	}

	var results []*TickerData
	for _, data := range resp.Data.MarketChart {
		result := NewTickerData(time.UnixMilli(data.Timestamp), data.Price)
		results = append(results, result)
	}

	return results, nil
}

func handleHour4DataFetching(credentials *TokenInsightCredentials, tickerSymbol string, length int) ([]*TickerData, error) {
	lengthMultiplier := 4
	adjustedLength := lengthMultiplier * length
	interval := "hour"
	url := fmt.Sprintf("%s/%s?interval=%s&length=%d", credentials.baseUrl, strings.ToLower(tickerSymbol), interval, adjustedLength)

	resp, err := makeTokenInsightHistoricalDataCall(url, credentials.key)

	if err != nil {
		return nil, err
	}

	var results []*TickerData
	for i := len(resp.Data.MarketChart) - 1; len(results) < length; i -= lengthMultiplier {
		currRes := resp.Data.MarketChart[i]
		result := NewTickerData(time.UnixMilli(currRes.Timestamp), currRes.Price)
		results = append(results, result)
	}

	return results, nil
}

func handleWeek1DataFetching(credentials *TokenInsightCredentials, tickerSymbol string, length int) ([]*TickerData, error) {
	lengthMultiplier := 7

	// TODO find replacement for week 1
	adjustedLength := 365
	interval := "day"
	url := fmt.Sprintf("%s/%s?interval=%s&length=%d", credentials.baseUrl, strings.ToLower(tickerSymbol), interval, adjustedLength)

	resp, err := makeTokenInsightHistoricalDataCall(url, credentials.key)

	if err != nil {
		return nil, err
	}

	var results []*TickerData

	// TODO find replacement for week 1
	for i := len(resp.Data.MarketChart) - 1; len(results) < 52; i -= lengthMultiplier {
		currRes := resp.Data.MarketChart[i]
		result := NewTickerData(time.UnixMilli(currRes.Timestamp), currRes.Price)
		results = append(results, result)
	}

	return results, nil
}

func makeTokenInsightHistoricalDataCall(url, key string) (*TokenInsightDataResp, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("TI_API_KEY", key)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var resp TokenInsightDataResp

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
