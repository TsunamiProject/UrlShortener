package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
	"strconv"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseUrl         string `env:"BASE_URL" envDefault:"localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func New() *Config {
	var config Config
	//flag.Func("-a", "Server address", func(s string) error {
	//	hp := strings.Split(s, ":")
	//	if len(hp) != 2 {
	//		return errors.New("need address in a form host:port")
	//	}
	//	config.ServerAddress = s
	//	return nil
	//})

	//flag.Func("-b", "Base url of short url", func(s string) error {
	//
	//})
	//flag.Func("-f", "File storage path", f)
	//
	//flag.Var(config.BaseUrl, "b", "Base url of short url")
	//flag.Var(config.FileStoragePath, "f", "File storage path")
	//
	//flag.Parse()
	//fmt.Println(config.ServerAddress)
	//fmt.Println(config.BaseUrl)
	//fmt.Println(config.FileStoragePath)

	err := env.Parse(&config)
	if err != nil {
		log.Fatal("Err with collecting app config", err)
	}

	return &config
}

//return config param from .env if exists or default value or default
func getEnv(key string, defaultVal string) string {
	//search for a parameter in .env file
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("%s value not found in .env file\n", key)

	return defaultVal
}

//return the converted from string bool param from .env if exists or default
func getEnvBool(name string, defaultVal bool) bool {
	envVal := getEnv(name, "")
	if val, err := strconv.ParseBool(envVal); err == nil {
		return val
	}

	return defaultVal
}

//func getFlags() {
//	flag.Var(addr, "addr", "Net address host:port")
//	flag.Var(addr, "addr", "Net address host:port")
//	flag.Var(addr, "addr", "Net address host:port")
//
//}

//func (a Config) String() string {
//	return a.ServerAddress
//}
//
//func (a *Config) Set(s string) error {
//	hp := strings.Split(s, ":")
//	if len(hp) != 2 {
//		return errors.New("Need address in a form host:port")
//	}
//	a.ServerAddress = s
//	a.BaseUrl = hp[0]
//	a.FileStoragePath = hp[0]
//	//port, err := strconv.Atoi(hp[1])
//	//if err != nil {
//	//	return err
//	//}
//	//a.Host = hp[0]
//	//a.Port = port
//	return nil
//}
