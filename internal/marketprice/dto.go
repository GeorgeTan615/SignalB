package marketprice

type MetadataResp struct {
	Symbol   string
	Interval string
	Timezone string
}

type Result struct {
	Date     string
	Open     float64
	Close    float64
	High     float64
	Low      float64
	Volume   float64
	AdjClose float64
}

type RapidAPIDataResp struct {
	Metadata MetadataResp
	Results  []Result
}

type MarketChartResp struct {
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

type TIDataResp struct {
	Name        string            `json:"name"`
	Symbol      string            `json:"symbol"`
	MarketChart []MarketChartResp `json:"market_chart"`
}

type TokenInsightDataResp struct {
	Data TIDataResp `json:"data"`
}

type CoinAPIDataResp struct {
	TimePeriodEnd string  `json:"time_period_end"`
	PriceClose    float64 `json:"price_close"`
}

type RefreshPriceResp struct {
	Ticker          string        `json:"ticker"`
	Class           string        `json:"class"`
	Timeframe       string        `json:"timeframe"`
	RefreshedPrices []*TickerData `json:"refreshedPrices"`
}
