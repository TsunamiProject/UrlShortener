package main

import (
	"log"
	"net/http"

	"github.com/TsunamiProject/UrlShortener.git/internal/app"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
)

func main() {
	//Creating config instance
	log.Print("Initializing config")
	cfg := config.New()

	//Creating server instance
	r := app.NewRouter()

	log.Printf("Server started on %s with BaseURL param: %s with file s path: %s "+"and "+
		"DatabaseDSN string: %s", cfg.ServerAddress, cfg.BaseURL, cfg.FileStoragePath, cfg.DatabaseDSN)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
