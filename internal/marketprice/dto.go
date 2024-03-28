package marketprice

type (
	MetadataResp struct {
		Symbol   string `json:"Symbol"`
		Interval string `json:"Interval"`
		Timezone string `json:"Timezone"`
	}

	Result struct {
		Date     string  `json:"Date"`
		Open     float64 `json:"Open"`
		Close    float64 `json:"Close"`
		High     float64 `json:"High"`
		Low      float64 `json:"Low"`
		Volume   float64 `json:"Volume"`
		AdjClose float64 `json:"AdjClose"`
	}

	RapidAPIDataResp struct {
		Metadata MetadataResp `json:"Metadata"`
		Results  []Result     `json:"Results"`
	}

	MarketChartResp struct {
		Price     float64 `json:"price"`
		Timestamp int64   `json:"timestamp"`
	}

	TIDataResp struct {
		Name        string            `json:"name"`
		Symbol      string            `json:"symbol"`
		MarketChart []MarketChartResp `json:"market_chart"`
	}

	TokenInsightDataResp struct {
		Data TIDataResp `json:"data"`
	}

	CoinAPIDataResp struct {
		TimePeriodEnd string  `json:"time_period_end"`
		PriceClose    float64 `json:"price_close"`
	}

	RefreshPriceResp struct {
		Ticker          string        `json:"ticker"`
		Class           string        `json:"class"`
		Timeframe       string        `json:"timeframe"`
		RefreshedPrices []*TickerData `json:"refreshedPrices"`
	}
)
