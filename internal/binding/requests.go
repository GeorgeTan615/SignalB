package binding

type RegisterBindingReq struct {
	TickerSymbol string `json:"tickerSymbol"`
	Timeframe    string `json:"timeframe"`
	Strategy     string `json:"strategy"`
}
