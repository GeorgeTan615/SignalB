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
		c.JSON(http.StatusBadRequest, errors.NewErrorRespWithErr(errors.RequestDeserializationError, err))
		return
	}

	if !utils.SliceContains[string](strategy.AllowedStrategies[:], req.Strategy) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Strategy must be of %v", strategy.AllowedStrategies)))
		return
	}

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], req.Timeframe) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Timeframe must be of %v", timeframe.AllowedTimeframes)))
		return
	}

	err := insertBinding(c.Request.Context(), req.TickerSymbol, req.Timeframe, req.Strategy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Insert binding failure", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Binding of %+v inserted successfully", req),
	})
}

func insertBinding(c context.Context, tickerSymbol, timeframe, strategy string) error {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	query := `insert into binding (ticker_symbol, timeframe, strategy) values (?,?,?)`
	_, err := database.MySqlDB.ExecContext(ctx, query, tickerSymbol, timeframe, strategy)

	if err != nil {
		return err
	}

	return err
}

func GetBindingsForTickerController(c *gin.Context) {
	tickerSymbol := c.Param("ticker")

	results, err := getBindingsByTicker(c, tickerSymbol)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error getting bindings by ticker", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bindings": results,
	})
}

func getBindingsByTicker(c context.Context, tickerSymbol string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	var results []interface{}
	err := database.SupabaseDBClient.DB.From("binding").Select("timeframe", "strategy").Eq("ticker_symbol", tickerSymbol).ExecuteWithContext(ctx, &results)

	return results, err

}

func GetBindingsForTimeframeController(c *gin.Context) {
	timeframe := c.Param("timeframe")

	results, err := getBindingsByTimeframe(c, timeframe)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error getting bindings by timeframe", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bindings": results,
	})
}

func getBindingsByTimeframe(c context.Context, timeframe string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	var results []interface{}
	err := database.SupabaseDBClient.DB.From("binding").Select("ticker_symbol", "strategy").Eq("timeframe", timeframe).ExecuteWithContext(ctx, &results)

	return results, err
}
