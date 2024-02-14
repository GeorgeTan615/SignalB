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
		log.Fatalln("Failed to load .env file", err)
	}

	stockDF := getStockDataFetcher()

	// TODO get crypto df

	fetcherManager = NewFetcherManager(stockDF)
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
