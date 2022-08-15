package app

import (
	"github.com/go-chi/chi/v5"

	"github.com/TsunamiProject/UrlShortener.git/internal/handlers"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers/middleware"
)

// NewRouter return router instance with handlers and error
func NewRouter() chi.Router {
	//Collecting http.Server instance
	//server := &http.Server{
	//	Addr:    config.IPPort.IP + ":" + config.IPPort.PORT,
	//	Handler: http.HandlerFunc(handlers.ReqHandler),
	//}
	//Collecting router
	//
	router := chi.NewRouter()
	router.Use(middleware.GzipRespWriteHandler, middleware.GzipReqParseHandler)
	router.Use(middleware.CookieHandler)
	router.Get("/*", handlers.GetURLHandler)
	router.Get("/api/user/urls", handlers.GetApiUserURLHandler)
	router.Post("/", handlers.ShortenerHandler)
	router.Post("/api/shorten", handlers.ShortenAPIHandler)
	router.Put("/{}", handlers.MethodNotAllowedHandler)
	router.Patch("/{}", handlers.MethodNotAllowedHandler)

	return router

}
