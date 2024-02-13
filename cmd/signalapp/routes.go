package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/binding"
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

	strategies := router.Group("/api/strategies")
	{
		strategies.GET("", strategy.GetStrategiesController)
	}

	bindings := router.Group("/api/bindings")
	{
		bindings.POST("", binding.RegisterBindingController)
		bindings.GET("/tickers/:ticker", binding.GetBindingsForTickerController)
		bindings.GET("/timeframes/:timeframe", binding.GetBindingsForTimeframeController)
	}

	// data := router.Group("/api/data")
	// {
	// 	data.POST("/:ticker/:timeframe", RefreshTickerTimeframeDataController)
	// 	data.POST("/:timeframe", RefreshTimeframeDataController)
	// 	data.GET("/:ticker/:timeframe", GetTickerTimeframeDataController)
	// }

}
