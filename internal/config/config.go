package config

import (
	"log"
	"os"
	"net/url"
	"errors"
	"github.com/spf13/pflag"
	"io/ioutil"
	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultBaseURL         = "http://localhost:8080"
	defaultFileStoragePath = "/tmp/test"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

//func New() *Config {
//	var config Config
//	err := env.Parse(&config)
//	if err != nil {
//		log.Fatal("Error with collecting env params", err)
//	}
//
//	if len(config.ServerAddress) > 0 && len(config.BaseURL) > 0 {
//		err = validateConfig(&config)
//		if err != nil {
//			log.Println(err)
//		}
//	}
//
//	parseFlags(&config)
//
//	if len(config.ServerAddress) > 0 && len(config.BaseURL) > 0 {
//		err = validateConfig(&config)
//		if err != nil {
//			log.Println(err)
//		}
//	} else {
//		log.Println("Configuration params not found. Collecting config with default params.")
//		setDefaultConfig(&config)
//		return &config
//	}
//
//	return &config
//}

func New() *Config {
	var flagConfig Config
	var envConfig Config
	var resultConfig Config

	parseFlags(&flagConfig)

	err := env.Parse(&envConfig)
	if err != nil {
		log.Fatal("Error with collecting env params", err)
	}

	if len(envConfig.ServerAddress) > 0 {
		err = validateURL(envConfig.ServerAddress)
		if err != nil {
			log.Println("Invalid server address param from env")
		}
		resultConfig.ServerAddress = envConfig.ServerAddress
	} else if len(flagConfig.ServerAddress) > 0 {
		if err != nil {
			log.Println("Invalid server address param from flag value")
		}
		resultConfig.ServerAddress = flagConfig.ServerAddress
	} else {
		log.Println("Server address param not found. Collecting config with default value.")
		resultConfig.ServerAddress = defaultServerAddress
	}

	if len(envConfig.BaseURL) > 0 {
		err = validateURL(envConfig.BaseURL)
		if err != nil {
			log.Println("Invalid base url param from env")
		}
		resultConfig.BaseURL = envConfig.BaseURL
	} else if len(flagConfig.BaseURL) > 0 {
		err = validateURL(flagConfig.BaseURL)
		if err != nil {
			log.Println("Invalid base url param from flag value")
		}
		resultConfig.BaseURL = flagConfig.BaseURL
	} else {
		log.Println("Base url param not found. Collecting config with default value.")
		resultConfig.BaseURL = defaultBaseURL
	}

	if len(envConfig.FileStoragePath) > 0 {
		err = validateFilePath(envConfig.FileStoragePath)
		if err != nil {
			log.Println("Invalid file storage path param from env")
		}
		resultConfig.FileStoragePath = envConfig.FileStoragePath
	} else if len(flagConfig.FileStoragePath) > 0 {
		err = validateFilePath(flagConfig.FileStoragePath)
		if err != nil {
			log.Println("Invalid file storage path param from flag value")
		}
		resultConfig.FileStoragePath = flagConfig.FileStoragePath
	} else {
		log.Println("File storage path param not found. Collecting config with default value.")
		resultConfig.FileStoragePath = defaultFileStoragePath
	}

	return &resultConfig
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
		return nil
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
	fileStorageErr := validateFilePath(c.FileStoragePath)
	if addrErr != nil {
		return errors.New("wrong server address param")
	}
	if baseURLErr != nil {
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

func setDefaultConfig(c *Config) {
	c.ServerAddress = defaultServerAddress
	c.BaseURL = defaultBaseURL
	c.FileStoragePath = defaultFileStoragePath
}
