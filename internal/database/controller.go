package database

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PingController(c *gin.Context) {
	query := `select symbol from ticker limit 1`

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	//nolint:rowserrcheck
	res, err := Client.DB.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}

	defer res.Close()

	c.JSON(http.StatusOK, gin.H{
		"message": "database ping-ed successfully",
	})
}
