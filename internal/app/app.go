package app

import (
	"github.com/go-chi/chi/v5"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/middleware"
)

// NewRouter return router instance with handlers and error
func NewRouter(rh *handlers.RequestHandler) chi.Router {
	//Collecting router
	router := chi.NewRouter()
	router.Use(middleware.GzipRespWriteHandler, middleware.GzipReqParseHandler)
	router.Use(middleware.CookieHandler)
	router.Get("/*", rh.GetURLHandler)
	router.Get("/api/user/urls", rh.GetAPIUserURLHandler)
	router.Get("/ping", rh.PingDBHandler)
	router.Post("/", rh.ShortenerHandler)
	router.Post("/api/shorten", rh.ShortenAPIHandler)
	router.Post("/api/shorten/batch", rh.ShortenAPIBatchHandler)
	router.Put("/{}", rh.MethodNotAllowedHandler)
	router.Patch("/{}", rh.MethodNotAllowedHandler)
	router.Delete("/api/user/urls", rh.DeleteURLHandler)

	return router
}
