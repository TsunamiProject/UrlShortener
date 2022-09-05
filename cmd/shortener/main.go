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

	var (
		stor  storage.Storage
		err   error
		dbObj *db.Database
	)

	switch {
	case cfg.DatabaseDSN != "":
		log.Println("Storage is DBStorage")
		dbObj, err = db.ConnectToDB(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err.Error())
		}
		stor, err = storage.GetDBStorage(cfg.BaseURL, dbObj)
		if err != nil {
			log.Fatal("failed to init dbSource: " + err.Error())
		}
		defer func(dbObj *db.Database) {
			err = dbObj.CloseDBConn()
			if err != nil {
				log.Fatal("failed to close db conn: " + err.Error())
			}
		}(dbObj)
	case cfg.FileStoragePath != "":
		log.Println("Storage is FileStorage")
		stor = storage.GetFileStorage(cfg.FileStoragePath, cfg.BaseURL)
	default:
		log.Println("Storage is InMemoryStorage")
		stor = storage.GetInMemoryStorage(cfg.BaseURL)
	}

	newHandler := handlers.NewRequestHandler(stor, dbObj)
	router := app.NewRouter(newHandler)

	log.Printf("Server started on %s with BaseURL param: %s with file s path: %s "+"and "+
		"DatabaseDSN string: %s", cfg.ServerAddress, cfg.BaseURL, cfg.FileStoragePath, cfg.DatabaseDSN)

	server := &http.Server{Addr: cfg.ServerAddress, Handler: router}
	log.Fatal(server.ListenAndServe())
}
