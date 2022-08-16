package storage

import "github.com/TsunamiProject/UrlShortener.git/internal/config"

var cfg = config.New()

type Storage interface {
	Read(url string) (string, int, error)
	Write(b []byte) (string, int, error)
	Restore() error
}
