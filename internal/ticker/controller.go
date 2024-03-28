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
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Errorf("%s: %w", errors.RequestDeserializationError, err)))
		return
	}

	if !utils.SliceContains[string](AllowedClasses[:], req.Class) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Errorf("valid classes: %s", AllowedClasses)))
		return
	}

	if err := insertTicker(c.Request.Context(), req.Symbol, req.Class); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp(fmt.Errorf("%s: %w", errors.DatabaseInsertionError, err)))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Ticker %s of class %s created successfully", req.Symbol, req.Class),
	})
}

func insertTicker(c context.Context, symbol, class string) error {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	return database.Client.InsertTicker(ctx, symbol, class)
}

func GetTickers(c *gin.Context) {
	results, err := getTickers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResp(fmt.Errorf("%s: %w", errors.DatabaseQueryError, err)))
	}

	c.JSON(http.StatusOK, gin.H{
		"tickers": results,
	})
}

func getTickers(c context.Context) ([]database.Ticker, error) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	return database.Client.GetTickers(ctx)
}
