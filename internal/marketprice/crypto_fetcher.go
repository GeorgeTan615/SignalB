package marketprice

import (
	"context"
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
	baseURL string
	key     string
}

type CoinAPICredentials struct {
	baseURL string
	key     string
}

type CryptoDataFetcher struct {
	tiCredentials        *TokenInsightCredentials
	coinAPICredentials   *CoinAPICredentials
	tickerToShorthandMap map[string]string
}

func NewCryptoDataFetcher(
	tiCredentials *TokenInsightCredentials,
	coinAPICredentials *CoinAPICredentials,
	tickerToShorthandMap map[string]string,
) *CryptoDataFetcher {
	return &CryptoDataFetcher{
		tiCredentials:        tiCredentials,
		coinAPICredentials:   coinAPICredentials,
		tickerToShorthandMap: tickerToShorthandMap,
	}
}

func (cryptoDF *CryptoDataFetcher) FetchClass() string {
	return ticker.CryptoClass
}

func (cryptoDF *CryptoDataFetcher) Fetch(
	ctx context.Context,
	timeframe,
	tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	if length > RefreshAllDataLength {
		return nil, fmt.Errorf("maximum length is %d", RefreshAllDataLength)
	}

	var (
		fetchStrategy func(context.Context, *CryptoDataFetcher, string, int) ([]*TickerData, error)
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

	return fetchStrategy(ctx, cryptoDF, tickerSymbol, length)
}

func handleDay1DataFetching(
	ctx context.Context,
	fetcher *CryptoDataFetcher,
	tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	interval := "day"
	url := fmt.Sprintf("%s/%s?interval=%s&length=%d", fetcher.tiCredentials.baseURL, strings.ToLower(tickerSymbol), interval, length)

	resp, err := makeTokenInsightHistoricalDataCall(ctx, url, fetcher.tiCredentials.key)
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

func handleHour4DataFetching(ctx context.Context, fetcher *CryptoDataFetcher, tickerSymbol string, length int) ([]*TickerData, error) {
	lengthMultiplier := 4
	adjustedLength := lengthMultiplier * length
	interval := "hour"
	url := fmt.Sprintf("%s/%s?interval=%s&length=%d", fetcher.tiCredentials.baseURL, strings.ToLower(tickerSymbol), interval, adjustedLength)

	resp, err := makeTokenInsightHistoricalDataCall(ctx, url, fetcher.tiCredentials.key)
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

// CoinAPI can't do concurrent calls with our tier, so can't have more than 1 crypto in W1.
func handleWeek1DataFetching(
	ctx context.Context,
	fetcher *CryptoDataFetcher,
	tickerSymbol string,
	length int,
) ([]*TickerData, error) {
	tickerShorthand, ok := fetcher.tickerToShorthandMap[tickerSymbol]

	if !ok {
		return nil, errors.New("cant map crypto ticker")
	}

	periodID := "7DAY"
	currTime := time.Now()
	timeEnd := currTime.Format("2006-01-02T15:04:05")
	timeStart := currTime.AddDate(0, 0, -length*7).Format("2006-01-02T15:04:05")

	url := fmt.Sprintf(
		"%s/BITSTAMP_SPOT_%s_USD/history?time_start=%s&time_end=%s&period_id=%s&limit=%d",
		fetcher.coinAPICredentials.baseURL,
		tickerShorthand,
		timeStart,
		timeEnd,
		periodID,
		length)

	resp, err := makeCoinAPIHistoricalDataCall(ctx, url, fetcher.coinAPICredentials.key)
	if err != nil {
		return nil, err
	}

	var results []*TickerData

	for i := len(resp) - 1; len(results) < length; i-- {
		currRes := resp[i]

		// Layout representing the format of the input string
		layout := "2006-01-02T15:04:05.9999999Z"

		// Parse the string to a time.Time value
		parsedTime, err := time.Parse(layout, currRes.TimePeriodEnd)
		if err != nil {
			return nil, err
		}

		result := NewTickerData(parsedTime, currRes.PriceClose)
		results = append(results, result)
	}

	return results, nil
}

func makeTokenInsightHistoricalDataCall(ctx context.Context, url, key string) (*TokenInsightDataResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func makeCoinAPIHistoricalDataCall(ctx context.Context, url, key string) ([]CoinAPIDataResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-CoinAPI-Key", key)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resp []CoinAPIDataResp

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
