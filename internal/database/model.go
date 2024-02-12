package database

type Ticker struct {
	Symbol string `json:"symbol"`
	Class  string `json:"class"`
}

func NewTicker(symbol, class string) *Ticker {
	return &Ticker{
		Symbol: symbol,
		Class:  class,
	}
}
