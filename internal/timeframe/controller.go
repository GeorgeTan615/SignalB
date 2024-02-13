package timeframe

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListTimeframesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"timeframes": AllowedTimeframes,
	})
}
