package config

import (
	"log"
	"os"
	"strconv"
)

type IpPortConfig struct {
	IP   string
	PORT string
}

type Config struct {
	Debug  bool
	IpPort IpPortConfig
}

func New() *Config {
	config := &Config{
		Debug: getEnvBool("DEBUG", false),
		IpPort: IpPortConfig{
			IP:   getEnv("IP", "localhost"),
			PORT: getEnv("PORT", "8080"),
		},
	}

	return config
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
