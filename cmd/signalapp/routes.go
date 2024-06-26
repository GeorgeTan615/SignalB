package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/binding"
	"github.com/signalb/internal/database"
	"github.com/signalb/internal/marketprice"
	"github.com/signalb/internal/strategy"
	"github.com/signalb/internal/ticker"
	"github.com/signalb/internal/timeframe"
)

func InitRoutes(router *gin.Engine) {
	timeframes := router.Group("/api/timeframes")
	{
		timeframes.GET("", timeframe.ListTimeframesController)
	}

	tickers := router.Group("/api/tickers")
	{
		tickers.POST("", ticker.RegisterTicker)
		tickers.GET("", ticker.GetTickers)
	}

	bindings := router.Group("/api/bindings")
	{
		bindings.POST("", binding.RegisterBindingController)
		bindings.GET("/tickers/:ticker", binding.GetBindingsForTickerController)
		bindings.GET("/timeframes/:timeframe", binding.GetBindingsForTimeframeController)
	}

	data := router.Group("/api/marketprice")
	{
		data.POST("/:timeframe/:ticker", marketprice.RefreshPriceByTickerTimeframeController)
		data.POST("/:timeframe", marketprice.RefreshMarketpriceByTimeframeController)
		data.GET("/:timeframe/:ticker", marketprice.GetMarketpriceDataByTickerTimeframeController)
	}

	strategies := router.Group("/api/strategies")
	{
		strategies.GET("", strategy.GetStrategiesController)
		strategies.GET("/:timeframe/evaluate", strategy.EvaluateTickerStrategiesByTimeframeController)
	}

	router.GET("/ping", database.PingController)
}
