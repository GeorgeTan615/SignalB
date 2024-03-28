package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var Client *DBClient

func InitDB() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load .env file", err)
	}

	db, err := sql.Open("libsql", os.Getenv("TURSO_URL")+os.Getenv("TURSO_TOKEN"))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping: %v", err)
	}

	Client = newDBClient(db)

	log.Println("Successfully connected to database!")
}
