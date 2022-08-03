package config

import (
	"log"
	"os"
	"github.com/caarlos0/env/v6"
	"net/url"
	"errors"
	"github.com/spf13/pflag"
	"io/ioutil"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func New() *Config {
	var config Config
	err := env.Parse(&config)
	if err != nil {
		log.Fatal("Error with collecting env params", err)
	}

	if len(config.ServerAddress) > 0 && len(config.ServerAddress) > 0 {
		err = validateConfig(&config)
		if err != nil {
			log.Fatal(err)
		}
		return &config
	}

	parseFlags(&config)

	if len(config.ServerAddress) > 0 && len(config.ServerAddress) > 0 {
		err = validateConfig(&config)
		if err != nil {
			log.Fatal(err)
		}
		return &config
	} else {
		log.Fatal("Configuration params not found")
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
	if filepath == "" {
		return nil
	}
	_, err := os.Stat(filepath)
	if err == nil {
		log.Println(err)
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
	baseUrlErr := validateURL(c.BaseURL)
	fileStorageErr := validateFilePath(c.FileStoragePath)
	if addrErr != nil {
		return errors.New("wrong server address param")
	}
	if baseUrlErr != nil {
		return errors.New("wrong base url param")
	}
	if fileStorageErr != nil {
		return errors.New("wrong file storage path (dir doesnt exist)")
	}
	return nil
}

//collect config struct with flag values
func parseFlags(c *Config) {
	flagSet := pflag.FlagSet{}
	addrFlag := flagSet.StringP("-addr", "a", "", "Server address with format: host:port")
	baseURLFlag := flagSet.StringP("-baseurl", "b", "", "Base url of short urls")
	fileStorageFlag := flagSet.StringP("-fstorage", "f", "", "File storage path")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatal("Error while parsing sys Args")
	}

	c.ServerAddress = *addrFlag
	c.BaseURL = *baseURLFlag
	c.FileStoragePath = *fileStorageFlag
}
