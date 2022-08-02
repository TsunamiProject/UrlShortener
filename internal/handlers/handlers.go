package handlers

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
)

type Urls interface {
	Store(string, map[string]string) (string, error)
	Load(string) (string, error)
}

type UrlsTempStorage struct {
	Urls sync.Map
}
type EncodeStruct struct {
	Result string `json:"result"`
}
type DecodeStruct struct {
	Url string `json:"url"`
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
var cfg = config.New()

func RestoreFields() {
	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY, 0777)
		if err != nil {
			log.Printf("Error when opening file: %s", err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
			shortPart := strings.Split(scanner.Text(), cfg.BaseURL+"/")
			log.Println(string(shorten.DecodeString([]byte(shortPart[1]))))
			log.Println(shortPart[1])
			err = shortUrls.Store(shortPart[1],
				map[string]string{string(shorten.DecodeString([]byte(shortPart[1]))): shortPart[1]})
			if err != nil {
				return
			}

		}

		if err = scanner.Err(); err != nil {
			log.Printf("Error while reading file: %s", err)
			return
		}

		err = file.Close()
		if err != nil {
			return
		}

	}
}

// MethodNotAllowedHandler send http error if request method not allowed
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	//checking request method and sending error response if request method != get/post
	http.Error(w, "Only get/post methods are allowed!", http.StatusBadRequest)
	log.Printf("%d: %s request was recieved", http.StatusBadRequest, r.Method)
}

// ShortenerHandler send shorten url from full url which received from request body
func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	RestoreFields()
	log.Printf("Recieved request with method: %s from: %s",
		r.Method, r.Host)
	//calls saveUrlHandler on POST method
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res, status, err := storeURL(b)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), status)
		return
	}

	err = r.Body.Close()
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func ShortenApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s",
		r.Method, r.Host)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, status, err := urlDecoder(b)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), status)
		return
	}

	result, status, err := urlEncoder(res)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), status)
		return
	}
	//setting headers
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(result)
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
	reqURL := r.URL.String()
	res, status, err := getFullURL(reqURL)
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
func storeURL(b []byte) (string, int, error) {
	urlsMap := make(map[string]string)

	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	k, v := string(b), shorten.EncodeString(b)
	urlsMap[k] = v

	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			return "", http.StatusInternalServerError, nil
		}
		_, err = file.Write([]byte(fmt.Sprintf("%s/%s\n", cfg.BaseURL, urlsMap[k])))
		if err != nil {
			return "", http.StatusInternalServerError, nil
		}
		err = file.Close()
		if err != nil {
			return "", http.StatusInternalServerError, nil
		}
		res := fmt.Sprintf("%s/%s", cfg.BaseURL, urlsMap[k])
		return res, http.StatusCreated, nil
	}

	shortUrls.Urls.Store(v, urlsMap)
	res := fmt.Sprintf("%s/%s", cfg.BaseURL, urlsMap[k])
	return res, http.StatusCreated, nil
}

//return original url by ID as URL param, status code and error
func getFullURL(url string) (string, int, error) {
	if len(url) <= 1 {
		return "", http.StatusBadRequest, errors.New("missing parameter: ID")
	}

	v, err := shortUrls.Load(url[1:])
	if err != nil {
		return v, http.StatusNotFound, err
	}

	return v, http.StatusTemporaryRedirect, nil
}

func urlDecoder(b []byte) (string, int, error) {
	var decodeStruct DecodeStruct

	err := json.Unmarshal(b, &decodeStruct)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid request body")
	}

	res, status, err := storeURL([]byte(decodeStruct.Url))
	if err != nil {
		return "", status, err
	}
	return res, http.StatusCreated, nil
}

func urlEncoder(r string) ([]byte, int, error) {
	encodeStruct := EncodeStruct{Result: r}
	res, err := json.Marshal(encodeStruct)
	if err != nil {
		return []byte(""), http.StatusBadRequest, err
	}

	return res, http.StatusCreated, nil
}
