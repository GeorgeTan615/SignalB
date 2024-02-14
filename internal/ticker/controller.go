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

	if err := insertTicker(c.Request.Context(), req.Symbol, req.Class); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr(errors.DatabaseInsertionError, err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Ticker %s of class %s created successfully", req.Symbol, req.Class),
	})
}

func insertTicker(c context.Context, symbol, class string) error {
	ctx, cancel := context.WithTimeout(c, 2*time.Second)
	defer cancel()

	query := `insert into ticker (symbol, class) values (?,?)`

	_, err := database.MySqlDB.ExecContext(ctx, query, symbol, class)

	if err != nil {
		return err
	}

	return nil
}

func GetTickers(c *gin.Context) {
	results, err := getTickers(c.Request.Context())

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr(errors.DatabaseQueryError, err))
	}

	c.JSON(http.StatusOK, gin.H{
		"tickers": results,
	})
}

func getTickers(c context.Context) ([]database.Ticker, error) {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	query := `select symbol, class from ticker`

	res, err := database.MySqlDB.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}

	defer res.Close()

	tickers := []database.Ticker{}
	for res.Next() {
		var ticker database.Ticker

		err := res.Scan(&ticker.Symbol, &ticker.Class)

		if err != nil {
			return nil, err
		}

		tickers = append(tickers, ticker)
	}

	return tickers, err
}
