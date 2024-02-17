package strategy

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/errors"
	"github.com/signalb/internal/timeframe"
	"github.com/signalb/utils"
)

func GetStrategiesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"strategies": AllowedStrategies,
	})
}

func EvaluateTickerStrategiesByTimeframeController(c *gin.Context) {
	tf := c.Param("timeframe")

	if !utils.SliceContains[string](timeframe.AllowedTimeframes[:], tf) {
		c.JSON(http.StatusBadRequest, errors.NewErrorResp(fmt.Sprintf("Timeframe must be of %v", timeframe.AllowedTimeframes)))
		return
	}

	res, err := evaluateTickersStrategiesByTimeframe(tf)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorRespWithErr("Error evaluating strategies for each ticker in the given timeframe", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": res,
	})
}
