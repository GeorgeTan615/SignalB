package main

import (
	"github.com/gin-gonic/gin"
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

	// bindings := router.Group("/api/bindings")
	// {
	// 	bindings.POST("", RegisterTickerToTimeframeBindingController)
	// 	bindings.GET("", ListTickerToTimeframeBindingsController)
	// }

	// data := router.Group("/api/data")
	// {
	// 	data.GET("/:ticker/:timeframe", GetTickerTimeframeDataController)
	// 	data.POST("/:ticker/:timeframe", RefreshTickerTimeframeDataController)
	// 	data.POST("/:timeframe", RefreshTimeframeDataController)
	// }

}
