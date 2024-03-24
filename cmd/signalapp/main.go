package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
	"github.com/signalb/internal/marketprice"
	"github.com/signalb/internal/strategy"
	"github.com/signalb/internal/telegram"
)

func main() {
	router := gin.Default()

	// setup
	InitRoutes(router)
	telegram.InitBot()
	strategy.InitStrategies()
	marketprice.InitFetchers()
	database.InitDB()
	defer database.Client.Close()

	if err := router.Run(":8080"); err != nil {
		log.Println(err)
	}
}
