package marketprice

var cryptoDF *CryptoDataFetcher

type CryptoDataFetcher struct {
}

func NewCryptoDataFetcher() *CryptoDataFetcher {
	return &CryptoDataFetcher{}
}

func init() {
	cryptoDF = NewCryptoDataFetcher()
}

func (cryptoDF *CryptoDataFetcher) Fetch(timeframe, tickerSymbol string, length int) ([]*TickerData, error) {
	return nil, nil
}
