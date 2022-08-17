package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

var _ Storage = &URLsWithAuth{}

func GetInMemoryStorage() *URLsWithAuth {
	return &URLsWithAuth{}
}

type URLsWithAuth struct {
	AuthURLsStorage sync.Map
}

type JSONURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (u *URLsWithAuth) Write(b []byte, authCookieValue string) (string, error) {
	if len(b) == 0 {
		return "", errors.New("request body is empty")
	}
	originalURL, shortURL := string(b), shorten.EncodeString(b)
	urlsMap := make(map[string][]string)
	urlsMapForMarshall := make(map[string][]string)
	res, _ := u.AuthURLsStorage.Load(shortURL)
	temp, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(temp, &urlsMapForMarshall)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(urlsMapForMarshall[originalURL]); i++ {
		urlsMap[originalURL] = append(urlsMap[originalURL],
			urlsMapForMarshall[originalURL][i])
	}

	u.AuthURLsStorage.Delete(shortURL)
	urlsMap[originalURL] = append(urlsMap[originalURL], authCookieValue)
	u.AuthURLsStorage.Store(shortURL, urlsMap)

	resp := fmt.Sprintf("%s/%s", cfg.BaseURL, shortURL)

	return resp, nil
}

func (u *URLsWithAuth) Read(shortURL string) (string, error) {
	//read without cookie
	res, ok := u.AuthURLsStorage.Load(shortURL)
	log.Println("Data after loading from memory: ", res)
	if !ok {
		log.Println("No data")
		return "", fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	temp, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	urlsWithAuthMapForMarshalling := make(map[string][]string)
	err = json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return "", err
	}
	var originalURL string
	for k := range urlsWithAuthMapForMarshalling {
		originalURL = k
	}

	return originalURL, nil
}

func (u *URLsWithAuth) ReadAll(authCookieValue string) (string, error) {
	rangeMap := make(map[any]any)
	u.AuthURLsStorage.Range(func(key, value any) bool {
		log.Printf("Write: key: %v -- value: %v", key, value)
		rangeMap[key] = value
		return true
	})

	if len(rangeMap) == 0 {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	var toMarshallList []JSONURL
	for short, v := range rangeMap {
		temp, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		urlsWithAuthMapForMarshalling := make(map[string][]string)
		err = json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
		if err != nil {
			return "", err
		}
		for origin, val := range urlsWithAuthMapForMarshalling {
			for authIDs := range val {
				if val[authIDs] == authCookieValue {
					toMarshallList = append(toMarshallList, JSONURL{
						ShortURL:    fmt.Sprintf("%s/%v", cfg.BaseURL, short),
						OriginalURL: origin,
					})
				}
			}

		}
	}

	if len(toMarshallList) == 0 {
		return "", fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	resp, err := json.Marshal(toMarshallList)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
