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

type RapidApiResp struct {
	Metadata MetadataResp
	Results  []Result
}

type RefreshPriceResp struct {
	Ticker          string        `json:"ticker"`
	Class           string        `json:"class"`
	Timeframe       string        `json:"timeframe"`
	RefreshedPrices []*TickerData `json:"refreshedPrices"`
}
