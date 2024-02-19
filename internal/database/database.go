package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var MySqlDB *sql.DB

func InitDB() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalln("Failed to load .env file", err)
	}

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	MySqlDB = db

	log.Println("Successfully connected to PlanetScale!")
}
