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

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load .env file", err)
	}

	stockDF := getStockDataFetcher()
	cryptoDF := getCryptoDataFetcher()

	fetcherManager = NewFetcherManager(stockDF, cryptoDF)
}

func getStockDataFetcher() TickerDataFetcher {
	rapidApiBaseUrl := os.Getenv("RAPID_API_BASE_URL")
	rapidApiKey := os.Getenv("RAPID_API_KEY")
	rapidApiHost := os.Getenv("RAPID_API_HOST")

	credentials := &RapidApiCredentials{
		baseUrl: rapidApiBaseUrl,
		key:     rapidApiKey,
		host:    rapidApiHost,
	}

	return NewStockDataFetcher(credentials)
}

func getCryptoDataFetcher() TickerDataFetcher {
	tiBaseUrl := os.Getenv("TI_BASE_URL")
	tiApiKey := os.Getenv("TI_API_KEY")
	coinApiBaseUrl := os.Getenv("COINAPI_BASE_URL")
	coinApiKey := os.Getenv("COINAPI_API_KEY")

	tiCredentials := &TokenInsightCredentials{
		baseUrl: tiBaseUrl,
		key:     tiApiKey,
	}

	coinApiCredentials := &CoinApiCredentials{
		baseUrl: coinApiBaseUrl,
		key:     coinApiKey,
	}

	tickerToShorthandMap := map[string]string{
		"BITCOIN":  "BTC",
		"ETHEREUM": "ETH",
	}

	return NewCryptoDataFetcher(tiCredentials, coinApiCredentials, tickerToShorthandMap)
}
