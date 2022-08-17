package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

var _ Storage = &DBStorage{}

type DBStorage struct {
	db *db.Database
}

func GetDBStorage() (*DBStorage, error) {
	dbObj := db.ConnectToDB(cfg.DatabaseDSN)
	err := dbObj.CreateAuthTable()
	if err != nil {
		return nil, err
	}
	err = dbObj.CreateURLsTable()
	if err != nil {
		return nil, err
	}

	return &DBStorage{db: dbObj}, nil
}

func (db *DBStorage) Write(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}
	urls := &JSONURL{
		ShortURL:    shorten.EncodeString(b),
		OriginalURL: string(b),
	}

	err := db.db.InsertRow(authCookieValue, urls.ShortURL, urls.OriginalURL)
	if err != nil {
		return "", err
	}

	shortenURL := fmt.Sprintf("%s/%s", cfg.BaseURL, urls.ShortURL)

	return shortenURL, nil
}

func (db *DBStorage) Read(shortURL string) (string, error) {
	row, err := db.db.GetURLRow(shortURL)
	if err != nil || row == "" {
		return "", fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return row, nil
}

func (db *DBStorage) ReadAll(authCookieValue string) (string, error) {
	rows, err := db.db.GetAllRows(authCookieValue)
	if err != nil {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}
	var urlsList []JSONURL
	var shortURL string
	var originalURL string
	for rows.Next() {
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			log.Fatal("here?", err)
			return "", err
		}
		urlsList = append(urlsList, JSONURL{
			ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, shortURL),
			OriginalURL: originalURL,
		})
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	err = rows.Err()
	if err != nil {
		log.Fatal("here?", err)
		return "", err
	}
	if len(urlsList) == 0 {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)

	}
	res, err := json.Marshal(urlsList)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
