package handlers

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"

	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
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
	URL string `json:"url"`
}

type UrlsWithAuth struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UrlsWithAuthTempStorage struct {
	UrlsStorage sync.Map
}

func init() {
	if err := RestoreFields(); err != nil {
		log.Println(err)
	}
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
var urlsWithAuth UrlsWithAuthTempStorage
var cfg = config.New()

//RestoreFields collecting Urls struct from storage file if exists
func RestoreFields() error {
	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0666)
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return err
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			log.Println(scanner.Text())
			tempURL, err := url.Parse(scanner.Text())
			if err != nil {
				log.Printf("Error while parsing url: %s", err)
				return err
			}
			shortPart := tempURL.RequestURI()[1:]
			err = shortUrls.Store(shortPart,
				map[string]string{string(shorten.DecodeString([]byte(shortPart))): shortPart})
			if err != nil {
				return err
			}
		}

		if err = scanner.Err(); err != nil {
			log.Printf("Error while reading storage file: %s", err)
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// MethodNotAllowedHandler send http error if request method not allowed
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	//checking request method and sending error response if request method != get/post
	http.Error(w, "Only get/post methods are allowed!", http.StatusBadRequest)
	log.Printf("%d: %s request was recieved", http.StatusBadRequest, r.Method)
}

// ShortenerHandler send shorten url from full url which received from request body
func ShortenerHandler(w http.ResponseWriter, r *http.Request) {
	//calls saveUrlHandler on POST method
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("Recieved request with method: %s from: %s with r.body: %s",
		r.Method, r.Host, string(b))
	err = r.Body.Close()
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

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		err = storeURLsWithAuth(b, authCookie.Value)
		if err != nil {
			log.Printf("Error: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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

// ShortenAPIHandler send shorten url with json format from full url which received from request body
func ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Recieved request with method: %s from: %s with r.body: %s",
		r.Method, r.Host, string(b))

	err = r.Body.Close()
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res, status, err := urlDecoder(b, authCookie.Value)
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

func GetApiUserURLHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s with ID_PARAM: %s",
		r.Method, r.Host, r.URL.String()[1:])

	authCookie, err := r.Cookie("auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
	res, err := urlsWithAuth.getUserURLs(authCookie.Value)
	if len(res) == 0 {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
	w.Header().Set("content-type", "application/json")
	//w.Header().Set("Location", res)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(res)
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}

}

func (u *UrlsWithAuthTempStorage) getUserURLs(authCookie string) ([]byte, error) {
	res, _ := urlsWithAuth.UrlsStorage.Load(authCookie)
	if res == nil {
		return []byte(""), errors.New("urls not found")
	}

	temp, _ := json.Marshal(res)
	urlsWithAuthMapForMarshalling := make(map[string][]map[string]string)
	err := json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return []byte(""), err
	}
	var toMarshallList []UrlsWithAuth
	for i := 0; i < len(urlsWithAuthMapForMarshalling[authCookie]); i++ {
		iter := reflect.ValueOf(urlsWithAuthMapForMarshalling[authCookie][i]).MapRange()
		tempUrlsWithAuth := UrlsWithAuth{
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

//storeURL return short url from original url which must be in request body, status code and error
func storeURL(b []byte) (string, int, error) {
	urlsMap := make(map[string]string)

	if len(b) == 0 {
		return "", http.StatusBadRequest, errors.New("request body is empty")
	}

	k, v := string(b), shorten.EncodeString(b)
	urlsMap[k] = v

	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
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
	}

	shortUrls.Urls.Store(v, urlsMap)
	res := fmt.Sprintf("%s/%s", cfg.BaseURL, urlsMap[k])
	return res, http.StatusCreated, nil
}

func storeURLsWithAuth(b []byte, authCookieValue string) error {
	urlsWithAuthMap := make(map[string][]map[string]string)
	urlsWithAuthMapForMarshalling := make(map[string][]map[string]string)
	res, _ := urlsWithAuth.UrlsStorage.Load(authCookieValue)
	temp, _ := json.Marshal(res)
	err := json.Unmarshal(temp, &urlsWithAuthMapForMarshalling)
	if err != nil {
		return err
	}

	if len(b) == 0 {
		return errors.New("request body is empty")
	}

	k, v := string(b), shorten.EncodeString(b)

	urlsWithAuth.UrlsStorage.Delete(authCookieValue)

	for i := 0; i < len(urlsWithAuthMapForMarshalling[authCookieValue]); i++ {
		urlsWithAuthMap[authCookieValue] = append(urlsWithAuthMap[authCookieValue],
			urlsWithAuthMapForMarshalling[authCookieValue][i])
	}

	urlsWithAuthMap[authCookieValue] = append(urlsWithAuthMap[authCookieValue], map[string]string{k: v})
	urlsWithAuth.UrlsStorage.Store(authCookieValue, urlsWithAuthMap)
	return nil
}

//getFullURL return original url by ID as URL param, status code and error
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

func urlDecoder(b []byte, cookieValue string) (string, int, error) {
	var decodeStruct DecodeStruct

	err := json.Unmarshal(b, &decodeStruct)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid request body")
	}

	res, status, err := storeURL([]byte(decodeStruct.URL))
	if err != nil {
		return "", status, err
	}
	err = storeURLsWithAuth([]byte(decodeStruct.URL), cookieValue)
	if err != nil {
		return "", http.StatusInternalServerError, err
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
