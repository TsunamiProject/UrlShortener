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

func TestReqHandler(t *testing.T) {
	cfg := config.New()
	type want struct {
		statusCode  int
		response    string
		contentType string
		location    string
	}
	tests := []struct {
		name        string
		request     string
		requestBody string
		method      string
		want        want
	}{
		{
			name:        "Send request with no allowed method (PUT)",
			request:     "/",
			requestBody: "",
			method:      "PUT",
			want: want{
				statusCode:  400,
				response:    "Only get/post methods are allowed!\n",
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:        "Get shorten URL from origin URL. Request body is not empty.",
			request:     "/",
			requestBody: "http://endxivm.com/y1ryfyoiudul7",
			method:      "POST",
			want: want{
				statusCode:  201,
				response:    fmt.Sprintf("http://%s:%s/http://endxivm.com/y1ryf", cfg.IPPort.IP, cfg.IPPort.PORT),
				contentType: "application/json",
				location:    "",
			},
		},
		{
			name:        "Get origin URL from shorten URL. Shorten URL already exists.",
			request:     "/http://endxivm.com/y1ryf",
			requestBody: "",
			method:      "GET",
			want: want{
				statusCode:  307,
				response:    "",
				contentType: "application/json",
				location:    "http://endxivm.com/y1ryfyoiudul7",
			},
		},
		{
			name:        "Get shorten URL from origin URL. Request body is empty.",
			request:     "/",
			requestBody: "",
			method:      "POST",
			want: want{
				statusCode:  400,
				response:    "request body is empty\n",
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
		{
			name:        "Get origin URL from shorten URL. Shorten URL doesn't exist.",
			request:     "/http://endxivm.com/",
			requestBody: "",
			method:      "GET",
			want: want{
				statusCode:  404,
				response:    "there are no URLs with ID: http://endxivm.com/\n",
				contentType: "text/plain; charset=utf-8",
				location:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(ReqHandler)
			h.ServeHTTP(w, req)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))

			respBody, err := ioutil.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.response, string(respBody))
		})
	}
}
