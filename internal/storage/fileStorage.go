package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

var _ Storage = &FileStorage{}

func GetFileStorage() *FileStorage {
	return &FileStorage{}
}

type FileStorage struct {
}

type FileStruct struct {
	CookieValue string
	URLs        JSONURL
}

//return short url from original url which must be in request body, status code and error
func (f *FileStorage) Write(b []byte, authCookieValue string, ctx context.Context) (string, int, error) {
	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}
	toFile := &FileStruct{
		CookieValue: authCookieValue,
		URLs: JSONURL{
			ShortURL:    shorten.EncodeString(b),
			OriginalURL: string(b),
		},
	}
	res, err := json.Marshal(toFile)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	_, err = file.Write([]byte(fmt.Sprintf("%s\n", res)))
	if err != nil {
		err = file.Close()
		if err != nil {
			return "", http.StatusInternalServerError, nil
		}
		return "", http.StatusInternalServerError, nil
	}

	err = file.Close()
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	shortenURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shorten.EncodeString(b))

	return shortenURL, http.StatusCreated, nil
}

//return original url by ID as URL param, status code and error
func (f *FileStorage) Read(shortURL string, authCookieValue string, ctx context.Context) (string, int, error) {
	if len(shortURL) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}
	scanner := bufio.NewScanner(file)

	var originalURL string
	for scanner.Scan() {
		var temp FileStruct
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		if temp.CookieValue == authCookieValue && temp.URLs.ShortURL == shortURL {
			originalURL = temp.URLs.OriginalURL
		}
	}

	if err = scanner.Err(); err != nil {
		return "", http.StatusInternalServerError, nil
	}

	err = file.Close()
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	if originalURL == "" {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return originalURL, http.StatusTemporaryRedirect, nil
}

func (f *FileStorage) ReadAll(authCookieValue string, ctx context.Context) (string, int, error) {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}
	scanner := bufio.NewScanner(file)

	var resList []JSONURL
	for scanner.Scan() {
		var temp FileStruct
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		if temp.CookieValue == authCookieValue {
			resList = append(resList, temp.URLs)
		}
	}

	if err = scanner.Err(); err != nil {
		return "", http.StatusInternalServerError, nil
	}

	err = file.Close()
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	if len(resList) == 0 {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	resp, err := json.Marshal(resList)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	return string(resp), http.StatusOK, nil
}
