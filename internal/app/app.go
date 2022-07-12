package app

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers"
	"github.com/go-chi/chi/v5"
)

// NewRouter return router instance with handlers and error
func NewRouter() chi.Router {
	//Collecting http.Server instance
	//server := &http.Server{
	//	Addr:    config.IPPort.IP + ":" + config.IPPort.PORT,
	//	Handler: http.HandlerFunc(handlers.ReqHandler),
	//}
	//Collecting router
	router := chi.NewRouter()
	router.Get("/*", handlers.GetURLHandler)
	router.Post("/", handlers.ShortenerHandler)
	router.Put("/{}", handlers.MethodNotAllowedHandler)
	router.Patch("/{}", handlers.MethodNotAllowedHandler)

	return router

}
