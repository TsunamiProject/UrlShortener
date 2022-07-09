package main

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/app"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/joho/godotenv"
	"log"
)

func init() {
	//Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found. Starting with default config settings...")
	}
}

func main() {
	//Creating config instance
	log.Print("Initializing config")
	cfg := config.New()

	//Creating server instance
	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatal("Unable to create server instance with current config settings")
	}

	log.Printf("Server started on %s:%s. Debug mode is: %v", cfg.IPPort.IP, cfg.IPPort.PORT, cfg.Debug)
	log.Fatal(server.ListenAndServe())
}
