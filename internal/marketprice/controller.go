package marketprice

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/timeframe"
	"github.com/signalb/utils"
)

func RefreshMarketpriceByTickerTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")
	ticker := c.Param("ticker")
	reqCtx := c.Request.Context()

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Timeframe must be of %v", timeframe.AllowedTimeframes)))
		return
	}

	res, err := refreshPriceByTickerTimeframe(reqCtx, ticker, tf)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error refreshing data", err))
		return
	}

	c.JSON(http.StatusOK, res)
}

func RefreshMarketpriceByTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Timeframe must be of %v", timeframe.AllowedTimeframes)))
		return
	}

	reqCtx := c.Request.Context()

	res, err := refreshPriceByTimeframe(reqCtx, tf)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error refreshing data by timeframe", err))
		return
	}

	c.JSON(http.StatusOK, res)
}

func GetMarketpriceDataByTickerTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")
	ticker := c.Param("ticker")
	ctx := c.Request.Context()

	class, err := getTickerClass(ctx, ticker)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error getting ticker class", err))
		return
	}

	fetcher, ok := fetcherManager.getFetcherByTickerClass(class)

	if !ok {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp("Error getting data fetcher"))
		return
	}

	res, err := fetcher.Fetch(tf, ticker, RefreshAllDataLength)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error fetching data", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}
