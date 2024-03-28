package database

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PingController(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	err := Client.Ping(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "database ping-ed successfully",
	})
}
