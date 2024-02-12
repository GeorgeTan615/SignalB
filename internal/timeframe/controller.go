package timeframe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/signalb/constants"
)

func ListTimeframesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"timeframes": constants.Timeframes,
	})
}
