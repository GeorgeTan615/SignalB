package main

import "github.com/gin-gonic/gin"

func InitRoutes(router *gin.Engine) {

	router.GET("timeframes", ListTimeframesController)

	tickerToTimeframe := router.Group("/api/ticker-timeframe")
	{
		tickerToTimeframe.POST("", RegisterTickerToTimeframeController)
		tickerToTimeframe.GET("", ListTickerToTimeframeController)
	}

	tickerData := router.Group("/api/data")
	{
		tickerData.GET("/:ticker/:timeframe", GetTickerTimeframeDataController)
		tickerData.POST("/:ticker/:timeframe", RefreshTickerTimeframeDataController)
		tickerData.POST("/:timeframe", RefreshTimeframeDataController)
	}

}
