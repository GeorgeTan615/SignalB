package main

import (
	"github.com/gin-gonic/gin"
	"github.com/signalb/internal/database"
)

func main() {
	database.InitDatabase()
	router := gin.Default()
	InitRoutes(router)
}
