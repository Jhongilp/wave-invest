package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	PolygonAPIKey  string        `mapstructure:"POLYGON_API_KEY"`
	Port           string        `mapstructure:"PORT"`
	PollInterval   time.Duration `mapstructure:"POLL_INTERVAL_SECONDS"`
	AllowedOrigins []string      `mapstructure:"ALLOWED_ORIGINS"`
}

func Load() (*Config, error) {
	// Load .env file if present
	_ = godotenv.Load()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("POLL_INTERVAL_SECONDS", 30)
	viper.SetDefault("ALLOWED_ORIGINS", []string{"http://localhost:5173"})

	viper.AutomaticEnv()

	cfg := &Config{
		PolygonAPIKey:  viper.GetString("POLYGON_API_KEY"),
		Port:           viper.GetString("PORT"),
		PollInterval:   time.Duration(viper.GetInt("POLL_INTERVAL_SECONDS")) * time.Second,
		AllowedOrigins: viper.GetStringSlice("ALLOWED_ORIGINS"),
	}

	if cfg.PolygonAPIKey == "" {
		log.Println("WARNING: POLYGON_API_KEY not set. Using demo mode with mock data.")
	}

	return cfg, nil
}
