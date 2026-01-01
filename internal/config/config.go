package config

import (
	"os"
)

type Config struct {
	Port     string
	Mode     string
	LogLevel string
}

func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "8080"),
		Mode:     getEnv("MODE", "LIVE"),
		LogLevel: getEnv("LOG_LEVEL", "INFO"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
