package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/storage"
)

var cfg = config.New()
var currStorage storage.Storage
var authStorage = storage.GetInMemoryAuthStorage()

type EncodeStruct struct {
	Result string `json:"result"`
}
type DecodeStruct struct {
	URL string `json:"url"`
}

func init() {
	if cfg.FileStoragePath != "" {
		currStorage = storage.GetFileStorage()
	} else {
		currStorage = storage.GetInMemoryStorage()
	}

	//if cfg.DatabaseDSN != "" {
	//	currStorage = db.Database{}
	//}

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
	res, status, err := currStorage.Write(b)
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
		err = authStorage.Write(b, authCookie.Value)
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
	res, status, err := currStorage.Read(reqURL[1:])
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
	err := dbObj.Ping()
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
	res, err := authStorage.Read(authCookie.Value)
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

func urlDecoder(b []byte, cookieValue string) (string, int, error) {
	var decodeStruct DecodeStruct

	err := json.Unmarshal(b, &decodeStruct)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid request body")
	}

	res, status, err := currStorage.Write([]byte(decodeStruct.URL))
	if err != nil {
		return "", status, err
	}
	err = authStorage.Write([]byte(decodeStruct.URL), cookieValue)
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
