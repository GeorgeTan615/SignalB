package strategy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetStrategiesController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"strategies": AllowedStrategies,
	})
}
