package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
	"github.com/signalb/internal/telegram"
)

func main() {
	router := gin.Default()
	InitRoutes(router)
	telegram.InitBot()
	database.InitDB()
	defer database.MySqlDB.Close()

	router.Run("localhost:8181")
}
