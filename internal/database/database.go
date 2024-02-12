package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	supa "github.com/nedpals/supabase-go"
)

var SupabaseDBClient *supa.Client

func InitDatabase() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Failed to load .env file", err)
		return
	}
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase := supa.CreateClient(supabaseUrl, supabaseKey)
	SupabaseDBClient = supabase
}
