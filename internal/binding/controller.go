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
	// Check if ticker is registered
	checkCtx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	checkQuery := `select count(symbol) 
					from ticker 
					where symbol = ?`

	var count int
	err := database.Client.DB.QueryRowContext(checkCtx, checkQuery, tickerSymbol).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("%s is not registered", tickerSymbol)
	}

	// Register binding
	registerQuery := `insert into binding (ticker_symbol, timeframe, strategy) values (?,?,?)`
	_, err = database.Client.DB.ExecContext(checkCtx, registerQuery, tickerSymbol, timeframe, strategy)
	if err != nil {
		return err
	}

	return err
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
	query := `select ticker_symbol, timeframe, strategy
					from binding
					where ticker_symbol = ?`

	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	res, err := database.Client.DB.QueryContext(ctx, query, tickerSymbol)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var results []database.Binding
	for res.Next() {
		var binding database.Binding

		if err := res.Scan(&binding.Timeframe, &binding.Strategy, &binding.TickerSymbol); err != nil {
			return nil, err
		}
		results = append(results, binding)
	}

	return results, nil
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
	query := `select ticker_symbol, timeframe, strategy
					from binding
					where timeframe = ?`

	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	res, err := database.Client.DB.QueryContext(ctx, query, timeframe)
	if err != nil {
		return nil, err
	}

	defer res.Close()

	var results []database.Binding
	for res.Next() {
		var binding database.Binding

		if err = res.Scan(&binding.Timeframe, &binding.Strategy, &binding.TickerSymbol); err != nil {
			return nil, err
		}
		results = append(results, binding)
	}

	return results, nil
}
