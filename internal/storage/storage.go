package storage

import (
	"github.com/TsunamiProject/UrlShortener.git/internal/config"
)

var cfg = config.New()

type Storage interface {
	Read(url string) (string, error)
	Write(b []byte, cookieValue string) (string, error)
	ReadAll(authCookie string) (string, error)
	Batch(b []byte, cookieValue string) (string, error)
}

//int - смешение уровней абстракций - storage должен быть изолирован
