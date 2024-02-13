package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
)

func main() {
	router := gin.Default()
	InitRoutes(router)
	defer database.MySqlDB.Close()

	router.Run("localhost:8181")
}
