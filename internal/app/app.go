package app

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

//const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
//
//func RandStringBytes(n int) string {
//	b := make([]byte, n)
//	for i := range b {
//		b[i] = letterBytes[rand.Intn(len(letterBytes))]
//	}
//	return string(b)
//}

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

func reqHandler(w http.ResponseWriter, r *http.Request) {
	//checking request method and sending error response if request method != get/post
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Only get/post methods are allowed!", http.StatusBadRequest)
		log.Printf("%d: %s request was recieved", http.StatusBadRequest, r.Method)
		return
	}
	if r.Method == http.MethodGet {
		//calls getFullUrlHandler on GET method
		res, err := getFullURLHandler(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("Error: %s", err)
			return
		}
		//setting headers
		w.Header().Set("content-type", "application/json")
		w.Header().Set("Location", res)
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, err = w.Write([]byte(""))
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}
	}

	if r.Method == http.MethodPost {
		//calls saveUrlHandler on POST method
		res, err := saveURLHandler(r)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//setting headers
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(res))
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}

	}
}

//return short url from original url which must be in request body
func saveURLHandler(r *http.Request) (string, error) {
	urlsMap := make(map[string]string)
	b, err := io.ReadAll(r.Body)
	if err != nil || len(b) == 0 {
		return "", err
	}

	urlsMap[string(b)] = string(b[:len(b)-8])
	//rChars := RandStringBytes(24)
	shortUrls.Urls.Store(string(b[:len(b)-8]), urlsMap)

	return urlsMap[string(b)], err
}

//return original url by ID as URL param
func getFullURLHandler(r *http.Request) (string, error) {
	reqURL := r.URL.String()
	if len(reqURL) <= 1 {
		return "", errors.New("missing parameter: ID")
	}

	v, err := shortUrls.Load(reqURL[1:])
	if err != nil {
		return v, err
	}

	return v, nil
}

// NewServer return http.Server instances with config settings
func NewServer(config *config.Config) (*http.Server, error) {
	http.HandleFunc("/", reqHandler)

	//Collecting http.Server instance
	server := &http.Server{
		Addr: config.IPPort.IP + ":" + config.IPPort.PORT,
	}

	return server, nil

}
