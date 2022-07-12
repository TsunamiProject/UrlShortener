package handlers

import (
	"errors"
	"fmt"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"io"
	"log"
	"net/http"
	"reflect"
	"sync"
)

type Urls interface {
	Store(string, map[string]string) (string, error)
	Load(string) (string, error)
}

type UrlsTempStorage struct {
	Urls sync.Map
}

func (u *UrlsTempStorage) Store(key string, value map[string]string) error {
	if len(key) == 0 || len(value) == 0 {
		return errors.New("empty key/value")
	}

	u.Urls.Store(key, value)
	return nil
}

func (u *UrlsTempStorage) Load(key string) (string, error) {
	value, _ := u.Urls.Load(key)
	if value == nil {
		return "", fmt.Errorf("there are no URLs with ID: %s", key)
	}

	originURL := ""
	iter := reflect.ValueOf(value).MapRange()

	for iter.Next() {
		originURL = iter.Key().String()
	}

	return originURL, nil
}

var shortUrls UrlsTempStorage

// MethodNotAllowedHandler send http error if request method not allowed
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	//checking request method and sending error response if request method != get/post
	http.Error(w, "Only get/post methods are allowed!", http.StatusBadRequest)
	log.Printf("%d: %s request was recieved", http.StatusBadRequest, r.Method)
}

// ShortenerHandler send shorten url from full url which received from request body
func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s",
		r.Method, r.Host)
	//calls saveUrlHandler on POST method
	res, status, err := storeURL(r)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), status)
		return
	}
	//setting headers
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write([]byte(res))
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
}

// GetURLHandler send origin url by short url in "Location" header
func GetURLHandler(w http.ResponseWriter, r *http.Request) {
	//calls getFullUrlHandler on GET method
	log.Printf("Recieved request with method: %s from: %s with ID_PARAM: %s",
		r.Method, r.Host, r.URL.String()[1:])
	res, status, err := getFullURL(r)
	if err != nil {
		http.Error(w, err.Error(), status)
		log.Printf("Error: %s", err)
		return
	}
	//setting headers
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Location", res)
	w.WriteHeader(status)
	_, err = w.Write([]byte(""))
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
}

//return short url from original url which must be in request body, status code and error
func storeURL(r *http.Request) (string, int, error) {
	urlsMap := make(map[string]string)
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	k, v := string(b), string(b[:len(b)-8])
	urlsMap[k] = v

	shortUrls.Urls.Store(v, urlsMap)

	err = r.Body.Close()
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	cfg := config.New()
	res := fmt.Sprintf("http://%s:%s/%s", cfg.IPPort.IP, cfg.IPPort.PORT, urlsMap[k])

	return res, http.StatusCreated, err
}

//return original url by ID as URL param, status code and error
func getFullURL(r *http.Request) (string, int, error) {
	reqURL := r.URL.String()

	if len(reqURL) <= 1 {
		return "", http.StatusBadRequest, errors.New("missing parameter: ID")
	}

	v, err := shortUrls.Load(reqURL[1:])
	if err != nil {
		return v, http.StatusNotFound, err
	}

	return v, http.StatusTemporaryRedirect, nil
}
