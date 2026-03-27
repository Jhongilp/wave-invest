package config

import (
	"log"
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
	mu             sync.RWMutex
}

var (
	cfg  *Config
	once sync.Once
	// ModeChangeListeners is called when trading mode changes
	ModeChangeListeners []func()
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
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TradingMode == TradingModeDemo
}

// GetTradingMode returns the current trading mode
func (c *Config) GetTradingMode() TradingMode {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.TradingMode
}

// SetTradingMode changes the trading mode and updates eToro credentials
func (c *Config) SetTradingMode(mode TradingMode) {
	log.Printf("Config: SetTradingMode called with mode=%s", mode)
	c.mu.Lock()

	if mode != TradingModeDemo && mode != TradingModeReal {
		c.mu.Unlock()
		log.Printf("Config: Invalid mode %s, ignoring", mode)
		return
	}

	c.TradingMode = mode

	// Update eToro credentials based on new mode
	if mode == TradingModeDemo {
		c.EtoroAPIKey = os.Getenv("ETORO_DEMO_API_KEY")
		c.EtoroAPISecret = os.Getenv("ETORO_DEMO_API_SECRET")
	} else {
		c.EtoroAPIKey = os.Getenv("ETORO_API_KEY")
		c.EtoroAPISecret = os.Getenv("ETORO_API_SECRET")
	}

	log.Printf("Config: Mode changed to %s, notifying %d listeners", mode, len(ModeChangeListeners))

	// Release lock BEFORE notifying listeners to avoid deadlock
	// (listeners may call IsDemo() which needs the lock)
	c.mu.Unlock()

	// Notify listeners (e.g., eToro client to reinitialize)
	for i, listener := range ModeChangeListeners {
		log.Printf("Config: Calling listener %d", i)
		listener()
		log.Printf("Config: Listener %d completed", i)
	}
	log.Printf("Config: SetTradingMode completed")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
