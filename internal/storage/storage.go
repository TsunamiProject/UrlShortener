package storage

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
)

var cfg = config.New()

type Storage interface {
	Read(url string, cookieValue string) (string, int, error)
	Write(b []byte, cookieValue string) (string, int, error)
	ReadAll(authCookie string) (string, int, error)
}
