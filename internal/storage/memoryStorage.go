package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
)

var _ Storage = &URLs{}

func GetInMemoryStorage() *URLs {
	return &URLs{}
}

func GetInMemoryAuthStorage() *URLsWithAuth {
	return &URLsWithAuth{}
}

type URLs struct {
	URLsStorage sync.Map
}

type URLsWithAuth struct {
	AuthURLsStorage sync.Map
}

type JsonURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewMemStorage() (*URLs, *URLsWithAuth) {
	return &URLs{}, &URLsWithAuth{}
}

//return short url from original url which must be in request body, status code and error
func (u *URLs) Write(b []byte) (string, int, error) {
	urlsMap := make(map[string]string)

	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	k, v := string(b), shorten.EncodeString(b)
	urlsMap[k] = v

	u.URLsStorage.Store(v, urlsMap)
	res := fmt.Sprintf("%s/%s", cfg.BaseURL, urlsMap[k])
	return res, http.StatusCreated, nil
}

func (u *URLs) Restore() error {
	file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		log.Printf("Error when opening file: %s", err)
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var temp JsonURL
		err = json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			log.Println("Error file unmarshalling file row", err)
			return err
		}
		u.URLsStorage.Store(temp.ShortURL,
			map[string]string{temp.OriginalURL: temp.ShortURL})
	}

	if err = scanner.Err(); err != nil {
		log.Printf("Error while reading storage file: %s", err)
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

//return original url by ID as URL param, status code and error
func (u *URLs) Read(url string) (string, int, error) {
	if len(url) <= 1 {
		return "", http.StatusBadRequest, errors.New("missing parameter: ID")
	}

	v, ok := u.URLsStorage.Load(url)
	if ok == false {
		return "", http.StatusNotFound, fmt.Errorf("there are no URLs with ID: %s", url)
	}

	originURL := ""
	iter := reflect.ValueOf(v).MapRange()

	for iter.Next() {
		originURL = iter.Key().String()
	}

	return originURL, http.StatusTemporaryRedirect, nil
}

func (u *URLsWithAuth) Read(authCookie string) ([]byte, error) {
	res, _ := u.AuthURLsStorage.Load(authCookie)
	if res == nil {
		return []byte(""), errors.New("urls not found")
	}

	temp, _ := json.Marshal(res)
	urlsWithAuthMapForMarshalling := make(map[string][]map[string]string)
	err := json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return []byte(""), err
	}
	var toMarshallList []JsonURL
	for i := 0; i < len(urlsWithAuthMapForMarshalling[authCookie]); i++ {
		iter := reflect.ValueOf(urlsWithAuthMapForMarshalling[authCookie][i]).MapRange()
		tempUrlsWithAuth := JsonURL{
			ShortURL:    "",
			OriginalURL: "",
		}
		for iter.Next() {
			tempUrlsWithAuth.OriginalURL = iter.Key().String()
			tempUrlsWithAuth.ShortURL = fmt.Sprintf("%s/%s", cfg.BaseURL, iter.Value().String())
		}
		toMarshallList = append(toMarshallList, tempUrlsWithAuth)
	}
	resp, err := json.Marshal(toMarshallList)
	if err != nil {
		return []byte(""), err
	}

	return resp, nil
}

func (u *URLsWithAuth) Write(b []byte, authCookieValue string) error {
	urlsMap := make(map[string][]map[string]string)
	urlsMapForMarshall := make(map[string][]map[string]string)
	res, _ := u.AuthURLsStorage.Load(authCookieValue)
	temp, _ := json.Marshal(res)
	err := json.Unmarshal(temp, &urlsMapForMarshall)
	if err != nil {
		return err
	}

	if len(b) == 0 {
		return errors.New("request body is empty")
	}

	k, v := string(b), shorten.EncodeString(b)

	u.AuthURLsStorage.Delete(authCookieValue)

	for i := 0; i < len(urlsMapForMarshall[authCookieValue]); i++ {
		urlsMap[authCookieValue] = append(urlsMap[authCookieValue],
			urlsMapForMarshall[authCookieValue][i])
	}

	urlsMap[authCookieValue] = append(urlsMap[authCookieValue], map[string]string{k: v})
	u.AuthURLsStorage.Store(authCookieValue, urlsMap)
	return nil
}

func (u *URLsWithAuth) Restore() error {
	return nil
}
