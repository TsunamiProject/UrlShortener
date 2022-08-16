package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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

func (db *DBStorage) Write(b []byte, authCookieValue string, ctx context.Context) (string, int, error) {
	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}
	urls := &JSONURL{
		ShortURL:    shorten.EncodeString(b),
		OriginalURL: string(b),
	}

	err := db.db.InsertRow(authCookieValue, urls.ShortURL, urls.OriginalURL, ctx)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	shortenURL := fmt.Sprintf("%s/%s", cfg.BaseURL, urls.ShortURL)

	return shortenURL, http.StatusCreated, nil
}

func (db *DBStorage) Read(shortURL string, authCookieValue string, ctx context.Context) (string, int, error) {
	row, err := db.db.GetRow(shortURL, authCookieValue, ctx)
	if err != nil || row == "" {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return row, http.StatusTemporaryRedirect, nil
}

func (db *DBStorage) ReadAll(authCookieValue string, ctx context.Context) (string, int, error) {
	rows, err := db.db.GetAllRows(authCookieValue, ctx)
	if err != nil {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}
	var urlsList []JSONURL
	var shortURL string
	var originalURL string
	for rows.Next() {
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			log.Fatal("here?", err)
			return "", http.StatusInternalServerError, err
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
		return "", http.StatusInternalServerError, err
	}
	if len(urlsList) == 0 {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)

	}
	res, err := json.Marshal(urlsList)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	return string(res), http.StatusOK, nil
}
