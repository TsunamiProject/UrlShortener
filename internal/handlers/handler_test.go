package handlers

import (
	"fmt"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/shorten"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	firstTestURL  = "http://endxivm.com/y1ryfyoiudul7"
	secondTestURL = "http://test.com/"
)

//var cfg = config.New()

type want struct {
	statusCode  int
	response    string
	contentType string
	location    string
}
type tests struct {
	request     string
	requestBody string
	method      string
	want        want
}

func runTests(tm map[string]tests, t *testing.T) {
	for test, tfields := range tm {
		t.Run(test, func(t *testing.T) {
			req := httptest.NewRequest(tfields.method, tfields.request, strings.NewReader(tfields.requestBody))
			w := httptest.NewRecorder()
			switch {
			case tfields.method == "GET":
				h := http.HandlerFunc(GetURLHandler)
				h.ServeHTTP(w, req)
			case tfields.method == "POST":
				switch {
				case tfields.request == "/":
					h := http.HandlerFunc(ShortenerHandler)
					h.ServeHTTP(w, req)
				case tfields.request == "/api/shorten":
					h := http.HandlerFunc(ShortenAPIHandler)
					h.ServeHTTP(w, req)
				}
			case tfields.method == "PUT" || tfields.method == "DELETE":
				h := http.HandlerFunc(MethodNotAllowedHandler)
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

			assert.Equal(t, tfields.want.response, string(respBody))
		})
	}
}

func TestMethodNotAllowedHandler(t *testing.T) {
	testMap := make(map[string]tests)
	testMap["Send request with no allowed method (PUT)"] = tests{
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
	runTests(testMap, t)
}

func TestShortenerHandler(t *testing.T) {
	testMap := make(map[string]tests)
	hashString := shorten.EncodeString([]byte(firstTestURL))
	testMap["Make shorten URL from origin URL. Request body is not empty."] = tests{
		request:     "/",
		requestBody: firstTestURL,
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    fmt.Sprintf("%s/%s", cfg.BaseURL, hashString),
			contentType: "application/json",
			location:    "",
		},
	}
	testMap["Make shorten URL from origin URL. Request body is empty."] = tests{
		request:     "/",
		requestBody: "",
		method:      "POST",
		want: want{
			statusCode:  400,
			response:    "request body is empty\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	runTests(testMap, t)

}

func TestShortenerApiHandler(t *testing.T) {
	testMap := make(map[string]tests)
	testJSON := "{\"url\":\"http://test.com/y1ryfyoiu7\"}"
	testInvalidJSON := "{\"url\":\"http://endxivm.com/y1ry"
	testResponse := "{\"result\":\"http://localhost:8080/687474703a2f2f746573742e636f6d2f7931727966796f697537\"}"
	//fmt.Println(testJson, "  ", testInvalidJson, "  ", testResponse)
	testMap["Make shorten URL from origin URL with json response. Request body is not empty."] = tests{
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
	testMap["Make shorten URL from origin URL with json response. Request body is invalid."] = tests{
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
	runTests(testMap, t)

}

func TestGetUrlHandler(t *testing.T) {
	testMap := make(map[string]tests)
	firstHashString := shorten.EncodeString([]byte(firstTestURL))
	secondHashString := shorten.EncodeString([]byte(secondTestURL))
	testMap["Get origin URL from shorten URL. Shorten URL already exists."] = tests{
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
	testMap["Get origin URL from shorten URL. Shorten URL doesn't exist."] = tests{
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
	runTests(testMap, t)
}
