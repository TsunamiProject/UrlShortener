package storage

import (
	"context"

	"github.com/TsunamiProject/UrlShortener.git/internal/config"
)

var cfg = config.New()

type Storage interface {
	Read(url string, cookieValue string, ctx context.Context) (string, int, error)
	Write(b []byte, cookieValue string, ctx context.Context) (string, int, error)
	ReadAll(authCookie string, ctx context.Context) (string, int, error)
}
