package marketprice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/signalb/internal/timeframe"
)

type MetadataResp struct {
	Symbol   string
	Interval string
	Timezone string
}

type Result struct {
	Date     string
	Open     float64
	Close    float64
	High     float64
	Low      float64
	Volume   float64
	AdjClose float64
}

type RapidApiResp struct {
	Metadata MetadataResp
	Results  []Result
}

type RapidApiCredentials struct {
	baseUrl string
	key     string
	host    string
}

var stockDF *StockDataFetcher

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalln("Failed to load .env file", err)
	}
	rapidApiBaseUrl := os.Getenv("RAPID_API_BASE_URL")
	rapidApiKey := os.Getenv("RAPID_API_KEY")
	rapidApiHost := os.Getenv("RAPID_API_HOST")

	credentials := &RapidApiCredentials{
		baseUrl: rapidApiBaseUrl,
		key:     rapidApiKey,
		host:    rapidApiHost,
	}

	stockDF = NewStockDataFetcher(credentials)
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

func (stockDF *StockDataFetcher) Fetch(timeframe, tickerSymbol string, length int) ([]*TickerData, error) {
	if length >= 1000 {
		return nil, errors.New("maximum length is 1000")
	}

	var url string
	mappedTimeframeVal := stockDF.timeframeMapping[timeframe]
	if mappedTimeframeVal != "intraday" {
		location, err := time.LoadLocation("America/New_York")

		if err != nil {
			return nil, err
		}

		today := time.Now().In(location)
		dateEnd := fmt.Sprintf("%v-%v-%v", today.Year(), int(today.Month()), today.Day())

		yesterday := today.AddDate(0, 0, -length)
		dateStart := fmt.Sprintf("%v-%v-%v", yesterday.Year(), int(yesterday.Month()), yesterday.Day())

		url = fmt.Sprintf("%s/%s?symbol=%s&dateStart=%s&dateEnd=%s", stockDF.credentials.baseUrl, mappedTimeframeVal, tickerSymbol, dateStart, dateEnd)
	} else {
		adjustedLength := 4 * length
		url = fmt.Sprintf("%s/%s?symbol=%s&interval=60min&maxreturn=%d", stockDF.credentials.baseUrl, mappedTimeframeVal, tickerSymbol, adjustedLength)
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", stockDF.credentials.key)
	req.Header.Add("X-RapidAPI-Host", stockDF.credentials.host)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var resp RapidApiResp

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	fmt.Println(resp)
	return nil, nil
}
