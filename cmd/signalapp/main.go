package main

import (
	"log"

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

	if err := router.Run(":8080"); err != nil {
		log.Println(err)
	}
}
