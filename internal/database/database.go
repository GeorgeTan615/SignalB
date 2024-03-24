package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
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
