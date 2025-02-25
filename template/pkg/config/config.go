package config

import (
	"os"
)

type Config struct {
	AppPort string
}

func NewConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "7001"
	}

	return &Config{
		AppPort: port,
	}
}
