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

type DBStorage struct {
	db      *db.Database
	baseURL string
}

func GetDBStorage(baseURL string, dbObj *db.Database) (*DBStorage, error) {
	err := dbObj.CreateURLsTable()
	if err != nil {
		return nil, err
	}

	return &DBStorage{db: dbObj, baseURL: baseURL}, nil
}

func (dbObj *DBStorage) IsOk() error {
	err := dbObj.db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (dbObj *DBStorage) Write(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}
	urls := &JSONURL{
		ShortURL:    shorten.EncodeString(b),
		OriginalURL: string(b),
	}

	err := dbObj.db.InsertRow(authCookieValue, urls.ShortURL, urls.OriginalURL)
	if errors.Is(err, db.ErrDuplicateURL) {
		shortenURL := fmt.Sprintf("%s/%s", dbObj.baseURL, urls.ShortURL)
		return shortenURL, err
	}

	shortenURL := fmt.Sprintf("%s/%s", dbObj.baseURL, urls.ShortURL)

	return shortenURL, nil
}

func (dbObj *DBStorage) Read(shortURL string) (string, error) {
	row, err := dbObj.db.GetURLRow(shortURL)
	if errors.Is(err, db.ErrDeletedURL) {
		return "", err
	} else if err != nil || row == "" {
		return "", fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return row, nil
}

func (dbObj *DBStorage) ReadAll(authCookieValue string) (string, error) {
	rows, err := dbObj.db.GetAllRows(authCookieValue)
	if err != nil {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}
	var urlsList []JSONURL
	var shortURL string
	var originalURL string
	for rows.Next() {
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return "", err
		}
		urlsList = append(urlsList, JSONURL{
			ShortURL:    fmt.Sprintf("%s/%s", dbObj.baseURL, shortURL),
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

func (dbObj *DBStorage) Batch(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}

	var batchListBefore []db.BatchBefore
	err := json.Unmarshal(b, &batchListBefore)
	if err != nil {
		return "", err
	}
	batch, err := dbObj.db.Batch(batchListBefore, authCookieValue)
	if err != nil {
		return "", err
	}
	for i := range batch {
		batch[i].ShortURL = fmt.Sprintf("%s/%s", dbObj.baseURL, batch[i].ShortURL)
	}

	resp, err := json.Marshal(batch)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (dbObj *DBStorage) Delete(authCookieValue string, deleteList []string) error {
	if len(deleteList) == 0 {
		return errors.New("urls list to delete is empty")
	}

	err := dbObj.db.DeleteRows(authCookieValue, deleteList)
	if err != nil {
		return err
	}

	return nil
}
