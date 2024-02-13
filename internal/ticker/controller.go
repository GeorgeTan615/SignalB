package ticker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
	"github.com/signalb/internal/errors"
	"github.com/signalb/utils"
)

func RegisterTicker(c *gin.Context) {
	var req RegisterTickerReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorRespWithErr(errors.RequestDeserializationError, err))
		return
	}

	if !utils.SliceContains[string](AllowedClasses[:], req.Class) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Class can only be of %s", AllowedClasses)))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := insertTicker(ctx, req.Symbol, req.Class); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr(errors.DatabaseInsertionError, err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Ticker %s of class %s created successfully", req.Symbol, req.Class),
	})
}

func insertTicker(ctx context.Context, symbol, class string) error {
	entry := database.NewTicker(symbol, class)

	var results []database.Ticker
	err := database.SupabaseDBClient.DB.From("ticker").Insert(entry).ExecuteWithContext(ctx, &results)

	if err != nil {
		return err
	}

	return nil
}

func GetTickers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
	defer cancel()

	results, err := getTickers(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr(errors.DatabaseQueryError, err))
	}

	c.JSON(http.StatusOK, gin.H{
		"tickers": results,
	})
}

func getTickers(ctx context.Context) ([]database.Ticker, error) {
	var results []database.Ticker

	err := database.SupabaseDBClient.DB.From("ticker").Select("symbol", "class").ExecuteWithContext(ctx, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}
