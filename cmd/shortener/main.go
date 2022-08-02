package main

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/app"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func init() {
	//Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found. Starting with default config settings")
	}
}

func main() {
	//Creating config instance
	log.Print("Initializing config")
	cfg := config.New()

	//Creating server instance
	r := app.NewRouter()

	log.Printf("Server started on %s with BaseURL param: %s with file storage path: %s", cfg.ServerAddress,
		cfg.BaseURL, cfg.FileStoragePath)
	httpAddr := cfg.ServerAddress
	log.Fatal(http.ListenAndServe(httpAddr, r))
}
