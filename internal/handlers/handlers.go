package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/storage"
)

type RequestHandler struct {
	storage storage.Storage
	baseURL string
	dbDSN   string
}

func NewRequestHandler(storage storage.Storage, baseURL string, dbDSN string) *RequestHandler {
	return &RequestHandler{
		storage: storage,
		baseURL: baseURL,
		dbDSN:   dbDSN,
	}
}

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

// MethodNotAllowedHandler send http error if request method not allowed
func (rh *RequestHandler) MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	//checking request method and sending error response if request method != get/post
	http.Error(w, "Only get/post methods are allowed!", http.StatusBadRequest)
	log.Printf("%d: %s request was recieved", http.StatusBadRequest, r.Method)
}

// ShortenerHandler send shorten url from full url which received from request body
func (rh *RequestHandler) ShortenerHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	authCookie, err := r.Cookie("auth")
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	writeToStruct := &WriteTo{
		ReqBody:     b,
		CookieValue: authCookie.Value,
	}

	var statusCode int

	res, err := rh.storage.Write(writeToStruct.ReqBody, writeToStruct.CookieValue)
	if errors.Is(err, db.ErrDuplicateURL) {
		statusCode = http.StatusConflict
	} else if err == nil {
		statusCode = http.StatusCreated
	} else {
		log.Printf("Error: %s", err)
		statusCode = http.StatusInternalServerError
		http.Error(w, err.Error(), statusCode)
	}

	//setting headers
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write([]byte(res))
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}
}

// ShortenAPIHandler send shorten url with json format from full url which received from request body
func (rh *RequestHandler) ShortenAPIHandler(w http.ResponseWriter, r *http.Request) {
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

	res, statusCode, err := rh.urlDecoder(b, authCookie.Value)
	if err != nil {
		log.Printf("Error: %s", err)
		http.Error(w, err.Error(), statusCode)
		return
	}

	result, err := rh.urlEncoder(res)
	if err != nil {
		statusCode = http.StatusBadRequest
	}
	//setting headers
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(result)
	if err != nil {
		log.Printf("Error: %s", err)
		return
	}

}

// GetURLHandler send origin url by short url in "Location" header
func (rh *RequestHandler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	//calls getFullUrlHandler on GET method
	log.Printf("Recieved request with method: %s from: %s with ID_PARAM: %s",
		r.Method, r.Host, r.URL.String()[1:])
	//ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
	//defer cancel()

	reqURL := r.URL.String()
	res, err := rh.storage.Read(reqURL[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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

//PingDBHandler send 200 status code if db is available and 500 if not responding
func (rh *RequestHandler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s. Ping DB",
		r.Method, r.Host)

	dbObj := db.ConnectToDB(rh.dbDSN)
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
func (rh *RequestHandler) GetAPIUserURLHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Recieved request with method: %s from: %s with ID_PARAM: %s",
		r.Method, r.Host, r.URL.String()[1:])

	authCookie, err := r.Cookie("auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	}
	//ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
	//defer cancel()
	res, err := rh.storage.ReadAll(authCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(res))
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}
	}
}

func (rh *RequestHandler) ShortenAPIBatchHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res, err := rh.storage.Batch(b, authCookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(res))
		if err != nil {
			log.Printf("Error: %s", err)
			return
		}
	}
}

func (rh *RequestHandler) urlDecoder(b []byte, cookieValue string) (string, int, error) {
	var decodeStruct DecodeStruct

	var statusCode int
	err := json.Unmarshal(b, &decodeStruct)
	if err != nil {
		return "", http.StatusBadRequest, errors.New("invalid request body")
	}

	res, err := rh.storage.Write([]byte(decodeStruct.URL), cookieValue)
	if errors.Is(err, db.ErrDuplicateURL) {
		statusCode = http.StatusConflict
		return res, statusCode, nil
	} else if err == nil {
		statusCode = http.StatusCreated
	} else {
		log.Printf("Error: %s", err)
		statusCode = http.StatusInternalServerError
		return "", statusCode, err
	}

	return res, statusCode, nil
}

func (rh *RequestHandler) urlEncoder(r string) ([]byte, error) {
	encodeStruct := EncodeStruct{Result: r}
	res, err := json.Marshal(encodeStruct)
	if err != nil {
		return []byte(""), err
	}

	return res, nil
}
