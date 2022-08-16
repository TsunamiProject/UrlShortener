package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/storage"
)

var cfg = config.New()
var currStorage storage.Storage

type EncodeStruct struct {
	Result string `json:"result"`
}
type DecodeStruct struct {
	URL string `json:"url"`
}

type WriteTo struct {
	ReqBody     []byte
	CookieValue string
}

func init() {
	if cfg.FileStoragePath != "" {
		currStorage = storage.GetFileStorage()
	} else {
		currStorage = storage.GetInMemoryStorage()
	}

	if cfg.DatabaseDSN != "" {
		//err := currStorage
		var err error
		currStorage, err = storage.GetDBStorage()
		if err != nil {
			log.Fatal(err)
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
	//calls saveUrlHandler on POST method
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Printf("Recieved request with method: %s from: %s with r.body: %s",
		r.Method, r.Host, string(b))
	err = r.Body.Close()

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeToStruct := &WriteTo{
		ReqBody:     b,
		CookieValue: authCookie.Value,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	res, status, err := currStorage.Write(writeToStruct.ReqBody, writeToStruct.CookieValue, ctx)
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	res, status, err := urlDecoder(b, authCookie.Value, ctx)
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
	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	reqURL := r.URL.String()
	res, status, err := currStorage.Read(reqURL[1:], authCookie.Value, ctx)
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

//PingDBHandler send 200 status code if db is available and 500 if not responding
func PingDBHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s. Ping DB",
		r.Method, r.Host)

	dbObj := db.ConnectToDB(cfg.DatabaseDSN)
	defer func(dbObj *db.Database) {
		err := dbObj.CloseDBConn()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

		}
	}(dbObj)
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()
	err := dbObj.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(""))
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}

}

//GetAPIUserURLHandler send json with all created urls with request cookie
func GetAPIUserURLHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s with ID_PARAM: %s",
		r.Method, r.Host, r.URL.String()[1:])

	authCookie, err := r.Cookie("auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	res, status, err := currStorage.ReadAll(authCookie.Value, ctx)
	if err != nil {
		http.Error(w, err.Error(), status)
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write([]byte(res))
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}

}

func urlDecoder(b []byte, cookieValue string, ctx context.Context) (string, int, error) {
	var decodeStruct DecodeStruct

	err := json.Unmarshal(b, &decodeStruct)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid request body")
	}

	res, status, err := currStorage.Write([]byte(decodeStruct.URL), cookieValue, ctx)
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
