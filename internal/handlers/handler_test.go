package handlers

import (
	"fmt"
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	for test, tfileds := range tm {
		t.Run(test, func(t *testing.T) {
			req := httptest.NewRequest(tfileds.method, tfileds.request, strings.NewReader(tfileds.requestBody))
			w := httptest.NewRecorder()
			if tfileds.method == "GET" {
				h := http.HandlerFunc(GetURLHandler)
				h.ServeHTTP(w, req)
			} else if tfileds.method == "POST" {
				h := http.HandlerFunc(ShortenerHandler)
				h.ServeHTTP(w, req)
			} else {
				h := http.HandlerFunc(MethodNotAllowedHandler)
				h.ServeHTTP(w, req)
			}

			res := w.Result()

			assert.Equal(t, tfileds.want.statusCode, res.StatusCode)
			assert.Equal(t, tfileds.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tfileds.want.location, res.Header.Get("Location"))

			respBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tfileds.want.response, string(respBody))
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
	cfg := config.New()
	testMap := make(map[string]tests)
	testMap["Get shorten URL from origin URL. Request body is not empty."] = tests{
		request:     "/",
		requestBody: "http://endxivm.com/y1ryfyoiudul7",
		method:      "POST",
		want: want{
			statusCode:  201,
			response:    fmt.Sprintf("http://%s:%s/http://endxivm.com/y1ryf", cfg.IPPort.IP, cfg.IPPort.PORT),
			contentType: "application/json",
			location:    "",
		},
	}
	testMap["Get shorten URL from origin URL. Request body is empty."] = tests{
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

func TestGetUrlHandler(t *testing.T) {
	testMap := make(map[string]tests)
	testMap["Get origin URL from shorten URL. Shorten URL already exists."] = tests{
		request:     "/http://endxivm.com/y1ryf",
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  307,
			response:    "",
			contentType: "application/json",
			location:    "http://endxivm.com/y1ryfyoiudul7",
		},
	}
	testMap["Get origin URL from shorten URL. Shorten URL doesn't exist."] = tests{
		request:     "/http://endxivm.com/",
		requestBody: "",
		method:      "GET",
		want: want{
			statusCode:  404,
			response:    "there are no URLs with ID: http://endxivm.com/\n",
			contentType: "text/plain; charset=utf-8",
			location:    "",
		},
	}
	runTests(testMap, t)
}
