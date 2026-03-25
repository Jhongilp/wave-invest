package config

import (
	"os"
	"sync"
)

type TradingMode string

const (
	TradingModeDemo TradingMode = "demo"
	TradingModeReal TradingMode = "real"
)

type Config struct {
	Port           string
	TradingMode    TradingMode
	EtoroAPIKey    string
	EtoroAPISecret string
	GeminiAPIKey   string
}

var (
	cfg  *Config
	once sync.Once
)

// Load initializes and returns the configuration (singleton)
func Load() *Config {
	once.Do(func() {
		mode := TradingMode(getEnv("TRADING_MODE", "demo"))
		if mode != TradingModeDemo && mode != TradingModeReal {
			mode = TradingModeDemo
		}

		// Select eToro credentials based on mode
		var apiKey, apiSecret string
		if mode == TradingModeDemo {
			apiKey = os.Getenv("ETORO_DEMO_API_KEY")
			apiSecret = os.Getenv("ETORO_DEMO_API_SECRET")
		} else {
			apiKey = os.Getenv("ETORO_API_KEY")
			apiSecret = os.Getenv("ETORO_API_SECRET")
		}

		cfg = &Config{
			Port:           getEnv("PORT", "8080"),
			TradingMode:    mode,
			EtoroAPIKey:    apiKey,
			EtoroAPISecret: apiSecret,
			GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		}
	})
	return cfg
}

// Get returns the loaded config (must call Load first)
func Get() *Config {
	return cfg
}

// IsDemo returns true if running in demo mode
func (c *Config) IsDemo() bool {
	return c.TradingMode == TradingModeDemo
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
