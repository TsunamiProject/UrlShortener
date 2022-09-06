package config

import (
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "/tmp/test"
	defaultDatabaseDSN     = "user=pqgotest dbname=pqgotest sslmode=verify-full"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func New() *Config {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}
	flagSet := pflag.FlagSet{}
	addrFlag := flagSet.StringP("-addr", "a", config.ServerAddress, "Server address with format: host:port")
	urlFlag := flagSet.StringP("-baseurl", "b", config.BaseURL, "Base url of short urls")
	fileStorageFlag := flagSet.StringP("-fstorage", "f", config.FileStoragePath, "File storage path")
	dbDSNFlag := flagSet.StringP("-dbDsn", "d", config.DatabaseDSN, "Database DSN string")

	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatal("Error while parsing sys Args")
	}
	config.ServerAddress = *addrFlag
	config.BaseURL = *urlFlag
	config.FileStoragePath = *fileStorageFlag
	config.DatabaseDSN = *dbDSNFlag

	err = validateConfig(&config)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

//parse url and return nil if url is valid or error
func validateURL(s string) error {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return err
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return err
	}
	return nil
}

//parse filepath and return nil if exist and error if not
func validateFilePath(filepath string) error {
	if len(filepath) == 0 {
		return errors.New("empty filepath")
	}
	_, err := os.Stat(filepath)
	if err == nil {
		return err
	}

	//trying to create file by filepath
	//if not - dir doesn't exist
	var d []byte
	err = ioutil.WriteFile(filepath, d, 0644)
	if err == nil {
		os.Remove(filepath)
		return err
	}

	return err
}

func validateConfig(c *Config) error {
	addrErr := validateURL(c.ServerAddress)
	baseURLErr := validateURL(c.BaseURL)
	//fileStorageErr := validateFilePath(c.FileStoragePath)
	if addrErr != nil {
		return errors.New("wrong server address param")
	}
	if baseURLErr != nil {
		return errors.New("wrong base url param")
	}
	//if fileStorageErr != nil {
	//	return errors.New("wrong file storage path (dir doesnt exist)")
	//}
	return nil
}
