package config

import (
	"os"
)

type Config struct {
	Port           string
	EtoroAPIKey    string
	EtoroAPISecret string
	GeminiAPIKey   string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		EtoroAPIKey:    os.Getenv("ETORO_API_KEY"),
		EtoroAPISecret: os.Getenv("ETORO_API_SECRET"),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
