package binding

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/strategy"
	"github.com/signalb/internal/timeframe"
	"github.com/signalb/utils"
)

func RegisterBindingController(c *gin.Context) {
	var req RegisterBindingReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("%s: %w", errors.RequestDeserializationError, err)))
		return
	}

	if !utils.SliceContains[string](strategy.AllowedStrategies, req.Strategy) {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("valid strategies: %v", strategy.AllowedStrategies)))
		return
	}

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], req.Timeframe) {
		c.JSON(http.StatusBadRequest,
			errors.NewErrorResp(fmt.Errorf("valid timeframes: %v", timeframe.AllowedTimeframes)))
		return
	}

	err := insertBinding(c.Request.Context(), req.TickerSymbol, req.Timeframe, req.Strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			errors.NewErrorResp(fmt.Errorf("insert binding: %w", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("binding of %+v inserted successfully", req),
	})
}

func insertBinding(c context.Context, tickerSymbol, timeframe, strategy string) error {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	if !database.Client.IsTickerRegistered(ctx, tickerSymbol) {
		return fmt.Errorf("%s is not registered", tickerSymbol)
	}

	return database.Client.InsertBinding(ctx, tickerSymbol, timeframe, strategy)
}

func GetBindingsForTickerController(c *gin.Context) {
	tickerSymbol := c.Param("ticker")

	results, err := getBindingsByTicker(c, tickerSymbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp(fmt.Errorf("get bindings by ticker: %w", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bindings": results,
	})
}

func getBindingsByTicker(c context.Context, tickerSymbol string) ([]database.Binding, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	return database.Client.GetBindingsByTicker(ctx, tickerSymbol)
}

func GetBindingsForTimeframeController(c *gin.Context) {
	timeframe := c.Param("timeframe")

	results, err := getBindingsByTimeframe(c, timeframe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp(fmt.Errorf("get bindings by timeframe: %w", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bindings": results,
	})
}

func getBindingsByTimeframe(c context.Context, timeframe string) ([]database.Binding, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	return database.Client.GetBindingsByTimeframe(ctx, timeframe)
}
