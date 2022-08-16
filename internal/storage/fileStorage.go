package storage

import (
	"bufio"
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

//return original url by ID as URL param, status code and error
func (f *FileStorage) Read(shortURL string) (string, int, error) {
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
		var temp JsonURL
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		if temp.ShortURL == shortURL {
			originalURL = temp.OriginalURL
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

//return short url from original url which must be in request body, status code and error
func (f *FileStorage) Write(b []byte) (string, int, error) {
	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}
	res, err := json.Marshal(&JsonURL{shorten.EncodeString(b), string(b)})
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}
	_, err = file.Write([]byte(fmt.Sprintf("%s\n", res)))
	if err != nil {
		err = file.Close()
		return "", http.StatusInternalServerError, nil
	}

	err = file.Close()
	if err != nil {
		return "", http.StatusInternalServerError, nil
	}

	shortenURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shorten.EncodeString(b))

	return shortenURL, http.StatusCreated, nil
}

func (f *FileStorage) Restore() error {
	return nil
}
