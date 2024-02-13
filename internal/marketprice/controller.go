package marketprice

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RefreshMarketpriceByTickerTimeframePriceController(c *gin.Context) {
	timeframe := c.Param("timeframe")
	ticker := c.Param("ticker")

	_, _ = stockDF.Fetch(timeframe, ticker, 10)

	// based on ticker, fetch type to know how we gonna pull data
	// probably can do some strategy pattern here
	// stock repository
	// crypto repository

	// a DataRepository, should be able to take a timeframe, ticker, and how many data you want (length) and return length amount of records

	// if we dont have 200 rows
	// just populate 200 rows

	// if we have 200 rows
	// get 1 new one row and discard oldest 1 row

	c.JSON(http.StatusOK, gin.H{})

}
