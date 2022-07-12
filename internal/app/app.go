package app

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
	"github.com/TsunamiProject/UrlShortener.git/internal/handlers"
	"net/http"
)

// NewServer return http.Server instances with config settings
func NewServer(config *config.Config) (*http.Server, error) {
	//Collecting http.Server instance
	server := &http.Server{
		Addr:    config.IPPort.IP + ":" + config.IPPort.PORT,
		Handler: http.HandlerFunc(handlers.ReqHandler),
	}

	return server, nil

}
