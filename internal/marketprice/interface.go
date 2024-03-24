package marketprice

import (
	"context"
	"time"
)

type TickerData struct {
	Time  time.Time
	Price float64
}

func NewTickerData(time time.Time, price float64) *TickerData {
	return &TickerData{
		Time:  time,
		Price: price,
	}
}

type TickerDataFetcher interface {
	Fetch(ctx context.Context, timeframe, ticker string, length int) ([]*TickerData, error)
	FetchClass() string
}
