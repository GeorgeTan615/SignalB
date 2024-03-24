package marketprice

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var fetcherManager *FetcherManager

type FetcherManager struct {
	ClassToFetcher map[string]TickerDataFetcher
}

func NewFetcherManager(fetchers ...TickerDataFetcher) *FetcherManager {
	classToFetcher := make(map[string]TickerDataFetcher)

	for _, fetcher := range fetchers {
		classToFetcher[fetcher.FetchClass()] = fetcher
	}

	return &FetcherManager{
		ClassToFetcher: classToFetcher,
	}
}

func (fm *FetcherManager) getFetcherByTickerClass(class string) (TickerDataFetcher, bool) {
	fetcher, ok := fm.ClassToFetcher[class]
	return fetcher, ok
}

func InitFetchers() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load .env file", err)
	}

	stockDF := getStockDataFetcher()
	cryptoDF := getCryptoDataFetcher()

	fetcherManager = NewFetcherManager(stockDF, cryptoDF)
}

func getStockDataFetcher() TickerDataFetcher {
	rapidAPIBaseURL := os.Getenv("RAPID_API_BASE_URL")
	rapidAPIKey := os.Getenv("RAPID_API_KEY")
	rapidAPIHost := os.Getenv("RAPID_API_HOST")

	credentials := &RapidAPICredentials{
		baseURL: rapidAPIBaseURL,
		key:     rapidAPIKey,
		host:    rapidAPIHost,
	}

	return NewStockDataFetcher(credentials)
}

func getCryptoDataFetcher() TickerDataFetcher {
	tiBaseURL := os.Getenv("TI_BASE_URL")
	tiAPIKey := os.Getenv("TI_API_KEY")
	coinAPIBaseURL := os.Getenv("COINAPI_BASE_URL")
	coinAPIKey := os.Getenv("COINAPI_API_KEY")

	tiCredentials := &TokenInsightCredentials{
		baseURL: tiBaseURL,
		key:     tiAPIKey,
	}

	coinAPICredentials := &CoinAPICredentials{
		baseURL: coinAPIBaseURL,
		key:     coinAPIKey,
	}

	tickerToShorthandMap := map[string]string{
		"BITCOIN":  "BTC",
		"ETHEREUM": "ETH",
	}

	return NewCryptoDataFetcher(tiCredentials, coinAPICredentials, tickerToShorthandMap)
}
