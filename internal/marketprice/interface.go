package marketprice

import "time"

type TickerData struct {
	Time  time.Time
	Price float64
}

type TickerDataFetcher interface {
	Fetch(timeframe, ticker string, length int) []*TickerData
}
