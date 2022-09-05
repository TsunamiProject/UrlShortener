package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/db"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
	"github.com/TsunamiProject/UrlShortener.git/internal/storage"
)

func runTestsInDB(s storage.Storage, db *db.Database, tm map[string]tests, t *testing.T) {
	if s == nil {
		t.Skip()
	}
	rh := NewRequestHandler(s, db)
	for test, tfields := range tm {
		t.Run(test, func(t *testing.T) {
			req := httptest.NewRequest(tfields.method, tfields.request, strings.NewReader(tfields.requestBody))
			w := httptest.NewRecorder()
			http.SetCookie(w, testCookie)
			req.AddCookie(testCookie)
			switch {
			case tfields.method == "GET":
				switch {
				case tfields.request != "/api/user/urls":
					h := http.HandlerFunc(rh.GetURLHandler)
					h.ServeHTTP(w, req)
				case tfields.request == "/api/user/urls":
					h := http.HandlerFunc(rh.GetAPIUserURLHandler)
					h.ServeHTTP(w, req)
				}
			case tfields.method == "POST":
				switch {
				case tfields.request == "/":
					h := http.HandlerFunc(rh.ShortenerHandler)
					h.ServeHTTP(w, req)
				case tfields.request == "/api/shorten":
					h := http.HandlerFunc(rh.ShortenAPIHandler)
					h.ServeHTTP(w, req)
				case tfields.request == "/api/shorten/batch":
					h := http.HandlerFunc(rh.ShortenAPIBatchHandler)
					h.ServeHTTP(w, req)
				}
			case tfields.method == "DELETE":
				h := http.HandlerFunc(rh.DeleteURLHandler)
				h.ServeHTTP(w, req)
			case tfields.method == "PUT":
				h := http.HandlerFunc(rh.MethodNotAllowedHandler)
				h.ServeHTTP(w, req)
			}

			res := w.Result()

			assert.Equal(t, tfields.want.statusCode, res.StatusCode)
			assert.Equal(t, tfields.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tfields.want.location, res.Header.Get("Location"))

			respBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)
			if tfields.request == "/api/user/urls" {
				log.Println(string(respBody))
				assert.Equal(t, len(tfields.want.response), len(string(respBody)))
			} else {
				assert.Equal(t, tfields.want.response, string(respBody))
			}
		})
	}
}

func TestMethodNotAllowedHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	testMap["#1 In DB: Send request with no allowed method (PUT)"] = tests{
		request:     "/",
		requestBody: "",
		method:      "PUT",
		want: want{
			statusCode:  400,
			response:    "Only get/post methods are allowed!\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestShortenerApiHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	testJSON := "{\"url\":\"http://test.com/\"}"
	testInvalidJSON := "{\"url\":\"http://endxivm.com/y1ry"
	testResponse := "{\"result\":\"http://localhost:8080/687474703a2f2f746573742e636f6d2f\"}"
	testMap["#1 In DB: Make shorten URL from origin URL with json response. Request body is not empty."] = tests{
		request:     "/api/shorten",
		requestBody: testJSON,
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    testResponse,
			contentType: "application/json",
			location:    "",
		},
	}
	testMap["#2 In DB: Make shorten URL from origin URL with json response. Request body is invalid."] = tests{
		request:     "/api/shorten",
		requestBody: testInvalidJSON,
		method:      "POST",
		want: want{
			statusCode:  400,
			response:    "invalid request body\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestShortenerHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	cfg := config.New()
	testMap := make(map[string]tests)
	hashStringFirstURL := shorten.EncodeString([]byte(firstTestURL))
	hashStringThirdURL := shorten.EncodeString([]byte(thirdTestURL))
	testMap["#1 In DB: Make shorten URL from origin URL. Request body is not empty."] = tests{
		request:     "/",
		requestBody: firstTestURL,
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    fmt.Sprintf("%s/%s", cfg.BaseURL, hashStringFirstURL),
			contentType: "application/json",
			location:    "",
		},
	}
	testMap["#2 In DB: Make shorten URL from origin URL. Request body is not empty."] = tests{
		request:     "/",
		requestBody: thirdTestURL,
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    fmt.Sprintf("%s/%s", cfg.BaseURL, hashStringThirdURL),
			contentType: "application/json",
			location:    "",
		},
	}
	testMap["#3 In DB: Make shorten URL from origin URL. Request body is empty."] = tests{
		request:     "/",
		requestBody: "",
		method:      "POST",
		want: want{
			statusCode:  500,
			response:    "request body is empty\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestGetUrlHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	firstHashString := shorten.EncodeString([]byte(firstTestURL))
	secondHashString := shorten.EncodeString([]byte("noexists"))
	testMap["#1 In DB: Get origin URL from shorten URL. Shorten URL already exists."] = tests{
		request:     "/" + firstHashString,
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  307,
			response:    "",
			contentType: "application/json",
			location:    firstTestURL,
		},
	}
	testMap["#2 In DB: Get origin URL from shorten URL. Shorten URL doesn't exist."] = tests{
		request:     "/" + secondHashString,
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  404,
			response:    fmt.Sprintf("there are no URLs with ID: %s\n", secondHashString),
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestGetUserUrlsHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	cfg := config.New()
	testMap := make(map[string]tests)
	firstHashString := shorten.EncodeString([]byte(firstTestURL))
	secondHashString := shorten.EncodeString([]byte(secondTestURL))
	thirdHashString := shorten.EncodeString([]byte(thirdTestURL))
	var testResSlice []storage.JSONURL
	testResSlice = append(testResSlice, storage.JSONURL{
		ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, secondHashString),
		OriginalURL: secondTestURL,
	}, storage.JSONURL{
		ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, firstHashString),
		OriginalURL: firstTestURL,
	}, storage.JSONURL{
		ShortURL:    fmt.Sprintf("%s/%s", cfg.BaseURL, thirdHashString),
		OriginalURL: thirdTestURL,
	})
	marshalledResSlice, _ := json.Marshal(testResSlice)
	testMap["#1 In DB: Get User urls by cookie"] = tests{
		request:     "/api/user/urls",
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  200,
			response:    string(marshalledResSlice),
			contentType: "application/json",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestShortenAPIBatchHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	var before []storage.BatchStructBefore
	before = append(before, storage.BatchStructBefore{
		CorrelationID: "testID",
		OriginalURL:   "http://rcteras2131kawj.net",
	})
	beforeMarshalled, _ := json.Marshal(before)
	log.Println("Before marshalled", string(beforeMarshalled))
	var after []storage.BatchStructAfter
	after = append(after, storage.BatchStructAfter{
		CorrelationID: "testID",
		ShortURL:      "http://localhost:8080/687474703a2f2f72637465726173323133316b61776a2e6e6574",
	})
	afterMarshalled, _ := json.Marshal(after)
	log.Println("After marshalled:", string(afterMarshalled))
	testMap["#1 In DB: Make batch of URLs. Request body is not empty."] = tests{
		request:     "/api/shorten/batch",
		requestBody: string(beforeMarshalled),
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    string(afterMarshalled),
			contentType: "application/json",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestDeleteHandlerInDB(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	firstHashString := shorten.EncodeString([]byte(firstTestURL))
	deleteList := make([]string, 0)
	deleteList = append(deleteList, firstHashString)
	testMap["#1 In DB: Delete url by User."] = tests{
		request:     "/api/user/urls",
		requestBody: fmt.Sprintf(`["%s"]`, deleteList[0]),
		method:      "DELETE",
		want: want{
			statusCode:  202,
			response:    "",
			contentType: "application/json",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}

func TestGetDeletedURL(t *testing.T) {
	dbConn, _ := db.ConnectToDB("postgres://shortener:pass@localhost:5432/shortener")
	testInDBStorage, _ := storage.GetDBStorage("http://localhost:8080", dbConn)
	defer dbConn.CloseDBConn()
	testMap := make(map[string]tests)
	firstHashString := shorten.EncodeString([]byte(firstTestURL))
	testMap["#2 In DB: Get deleted url by User."] = tests{
		request:     fmt.Sprintf("/%s", firstHashString),
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  410,
			response:    "URL is deleted\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	if testInDBStorage != nil {
		runTestsInDB(testInDBStorage, dbConn, testMap, t)
	}
}
