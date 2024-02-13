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

type Binding struct {
	TickerSymbol string `json:"ticker_symbol"`
	Timeframe    string `json:"timeframe"`
	Strategy     string `json:"strategy"`
}

func NewBinding(tickerSymbol, timeframe, strategy string) *Binding {
	return &Binding{
		TickerSymbol: tickerSymbol,
		Timeframe:    timeframe,
		Strategy:     strategy,
	}
}
