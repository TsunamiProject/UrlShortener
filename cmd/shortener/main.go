package main

import (
	"log"
	"net/http"

	"github.com/TsunamiProject/UrlShortener.git/internal/app"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers"
	"github.com/TsunamiProject/UrlShortener.git/internal/storage"
)

func main() {
	//Creating config instance
	log.Print("Initializing config")
	cfg := config.New()
	//
	////Creating server instance
	//r := app.NewRouter()
	//
	//log.Printf("Server started on %s with BaseURL param: %s with file s path: %s "+"and "+
	//	"DatabaseDSN string: %s", cfg.ServerAddress, cfg.BaseURL, cfg.FileStoragePath, cfg.DatabaseDSN)
	//log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
	var (
		stor  storage.Storage
		dbObj db.Database
		err   error
	)
	switch {
	case cfg.DatabaseDSN != "":
		log.Println("Storage is DBStorage")
		stor, err = storage.GetDBStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal("failed to init dbSource: " + err.Error())
		}
		defer func(dbObj *db.Database) {
			err = dbObj.CloseDBConn()
			if err != nil {
				log.Fatal(err)
			}
		}(&dbObj)
	case cfg.FileStoragePath != "":
		log.Println("Storage is FileStorage")
		stor = storage.GetFileStorage(cfg.FileStoragePath)
	default:
		log.Println("Storage is InMemoryStorage")
		stor = storage.GetInMemoryStorage()
	}

	newHandler := handlers.NewRequestHandler(stor, cfg.BaseURL, cfg.DatabaseDSN)
	router := app.NewRouter(newHandler)

	log.Printf("Server started on %s with BaseURL param: %s with file s path: %s "+"and "+
		"DatabaseDSN string: %s", cfg.ServerAddress, cfg.BaseURL, cfg.FileStoragePath, cfg.DatabaseDSN)

	server := &http.Server{Addr: cfg.ServerAddress, Handler: router}
	log.Fatal(server.ListenAndServe())
}
