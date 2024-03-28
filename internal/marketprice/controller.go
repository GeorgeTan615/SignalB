package marketprice

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/timeframe"
	"github.com/signalb/utils"
)

func RefreshPriceByTickerTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")
	ticker := c.Param("ticker")
	reqCtx := c.Request.Context()

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("valid timeframes: %v", timeframe.AllowedTimeframes)))
		return
	}

	res, err := refreshPriceByTickerTimeframe(reqCtx, ticker, tf)
	if err != nil {
		// c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error refreshing data", err))
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp(fmt.Errorf("error refreshing data: %w", err)))
		return
	}

	c.JSON(http.StatusOK, res)
}

func RefreshMarketpriceByTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("valid timeframes: %v", timeframe.AllowedTimeframes)))

		return
	}

	reqCtx := c.Request.Context()

	res, err := refreshPriceByTimeframe(reqCtx, tf)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("refresh data by timeframe: %w", err)))
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
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("get ticker class %w", err)))
		return
	}

	fetcher, ok := fetcherManager.getFetcherByTickerClass(class)

	if !ok {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("error getting data fetcher")))
		return
	}

	res, err := fetcher.Fetch(ctx, tf, ticker, RefreshAllDataLength)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("fetch data: %w", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}
