package database

import "time"

type Ticker struct {
	Symbol string `json:"symbol" db:"symbol"`
	Class  string `json:"class" db:"class"`
}

func NewTicker(symbol, class string) *Ticker {
	return &Ticker{
		Symbol: symbol,
		Class:  class,
	}
}

type Binding struct {
	TickerSymbol string `json:"ticker_symbol" db:"ticker_symbol"`
	Timeframe    string `json:"timeframe" db:"timeframe"`
	Strategy     string `json:"strategy" db:"strategy"`
}

func NewBinding(tickerSymbol, timeframe, strategy string) *Binding {
	return &Binding{
		TickerSymbol: tickerSymbol,
		Timeframe:    timeframe,
		Strategy:     strategy,
	}
}

type PriceData struct {
	TickerSymbol string
	Time         time.Time
	Price        float64
}
