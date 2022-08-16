package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

var _ Storage = &URLsWithAuth{}

func GetInMemoryStorage() *URLsWithAuth {
	return &URLsWithAuth{}
}

//func GetInMemoryAuthStorage() *URLsWithAuth {
//	return &URLsWithAuth{}
//}

//type URLs struct {
//	URLsStorage sync.Map
//}

type URLsWithAuth struct {
	AuthURLsStorage sync.Map
}

type JSONURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (u *URLsWithAuth) Write(b []byte, authCookieValue string, ctx context.Context) (string, int, error) {
	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	urlsMap := make(map[string][]map[string]string)
	urlsMapForMarshall := make(map[string][]map[string]string)
	res, _ := u.AuthURLsStorage.Load(authCookieValue)
	temp, _ := json.Marshal(res)
	err := json.Unmarshal(temp, &urlsMapForMarshall)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	k, v := string(b), shorten.EncodeString(b)

	u.AuthURLsStorage.Delete(authCookieValue)

	for i := 0; i < len(urlsMapForMarshall[authCookieValue]); i++ {
		urlsMap[authCookieValue] = append(urlsMap[authCookieValue],
			urlsMapForMarshall[authCookieValue][i])
	}

	urlsMap[authCookieValue] = append(urlsMap[authCookieValue], map[string]string{k: v})
	u.AuthURLsStorage.Store(authCookieValue, urlsMap)
	resp := fmt.Sprintf("%s/%s", cfg.BaseURL, v)

	return resp, http.StatusCreated, nil
}

func (u *URLsWithAuth) Read(shortURL string, authCookie string, ctx context.Context) (string, int, error) {
	res, _ := u.AuthURLsStorage.Load(authCookie)
	//if res == nil {
	//	log.Println("nil load ")
	//	return "", http.StatusNotFound, fmt.Errorf("there are no URLs with ID: %s", shortURL)
	//}

	temp, err := json.Marshal(res)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	urlsWithAuthMapForMarshalling := make(map[string][]map[string]string)
	err = json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	var originalURL string
	for i := 0; i < len(urlsWithAuthMapForMarshalling[authCookie]); i++ {
		iter := reflect.ValueOf(urlsWithAuthMapForMarshalling[authCookie][i]).MapRange()
		tempUrlsWithAuth := JSONURL{
			ShortURL:    "",
			OriginalURL: "",
		}
		for iter.Next() {
			tempUrlsWithAuth.OriginalURL = iter.Key().String()
			tempUrlsWithAuth.ShortURL = iter.Value().String()
			if tempUrlsWithAuth.ShortURL == shortURL {
				originalURL = tempUrlsWithAuth.OriginalURL
				break
			}
		}
	}
	if originalURL == "" {
		log.Println("Original URL not found")
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs with ID: %s", shortURL)
	}

	return originalURL, http.StatusTemporaryRedirect, nil
}

func (u *URLsWithAuth) ReadAll(authCookieValue string, ctx context.Context) (string, int, error) {
	res, _ := u.AuthURLsStorage.Load(authCookieValue)
	if res == nil {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	temp, _ := json.Marshal(res)
	urlsWithAuthMapForMarshalling := make(map[string][]map[string]string)
	err := json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	var toMarshallList []JSONURL
	for i := 0; i < len(urlsWithAuthMapForMarshalling[authCookieValue]); i++ {
		iter := reflect.ValueOf(urlsWithAuthMapForMarshalling[authCookieValue][i]).MapRange()
		tempUrlsWithAuth := JSONURL{
			ShortURL:    "",
			OriginalURL: "",
		}
		for iter.Next() {
			tempUrlsWithAuth.OriginalURL = iter.Key().String()
			tempUrlsWithAuth.ShortURL = fmt.Sprintf("%s/%s", cfg.BaseURL, iter.Value().String())
		}
		toMarshallList = append(toMarshallList, tempUrlsWithAuth)
	}

	if len(toMarshallList) == 0 {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs shortened by user: %s", authCookieValue)
	}

	resp, err := json.Marshal(toMarshallList)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return string(resp), http.StatusOK, nil
}
