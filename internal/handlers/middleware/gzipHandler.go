package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipRespWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipRespWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type gzipReqBodyReader struct {
	ReadCloser io.ReadCloser
}

func (g gzipReqBodyReader) Read(p []byte) (n int, err error) {
	return g.ReadCloser.Read(p)
}

func (g gzipReqBodyReader) Close() error {
	return g.ReadCloser.Close()
}

// GzipReqParseHandler check if headers contain Content-Encoding header, read request body with gzip Reader and passes
//to another handler
func GzipReqParseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var _ io.ReadCloser = gzipReqBodyReader{}
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer gzipReader.Close()
		r.Body = gzipReqBodyReader{gzipReader}
		next.ServeHTTP(w, r)
	})
}

// GzipRespWriteHandler check if headers contain Accept-Encoding header, write resp with gzip Writer and passes to
//another handler
func GzipRespWriteHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var _ http.ResponseWriter = gzipRespWriter{}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gzWriter, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			_, err = io.WriteString(w, err.Error())
			if err != nil {
				return
			}
			return
		}
		defer gzWriter.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipRespWriter{w, gzWriter}, r)
	})
}
