package main

import (
	api "api"
	db "db"
	"flag"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	var envPath string
	flag.StringVar(&envPath, "envPath", "./.env", "path to .env file (default: ./.env)")
	flag.Parse()

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("failed to load .env file at %q. \n go run . -envPath=/path/to/.env \nerr details: %v", envPath, err)
	}

	pgStore, err := db.NewPostgresStorage()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	apiPort := os.Getenv("API_PORT")
	server := api.NewAPIServer(apiPort, *pgStore)

	server.RunAPIServer()
}
