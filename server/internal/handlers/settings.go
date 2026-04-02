package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"wave_invest/config"
	"wave_invest/internal/services"
)

type TradingModeResponse struct {
	Mode   string `json:"mode"`
	IsDemo bool   `json:"isDemo"`
}

type SetTradingModeRequest struct {
	Mode string `json:"mode"`
}

// GetTradingMode returns the current trading mode (demo/real)
func GetTradingMode(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	mode := cfg.GetTradingMode()

	response := TradingModeResponse{
		Mode:   string(mode),
		IsDemo: mode == config.TradingModeDemo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SetTradingMode updates the trading mode (demo/real)
func SetTradingMode(w http.ResponseWriter, r *http.Request) {
	var req SetTradingModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var mode config.TradingMode
	switch req.Mode {
	case "demo":
		mode = config.TradingModeDemo
	case "real":
		mode = config.TradingModeReal
	default:
		http.Error(w, "Invalid mode. Must be 'demo' or 'real'", http.StatusBadRequest)
		return
	}

	// Persist to Firestore
	settingsService := services.NewSettingsService()
	if err := settingsService.SetTradingMode(r.Context(), string(mode)); err != nil {
		log.Printf("Failed to persist trading mode to Firestore: %v", err)
		// Continue anyway - will still work in memory
	}

	cfg := config.Get()
	cfg.SetTradingMode(mode)

	response := TradingModeResponse{
		Mode:   string(mode),
		IsDemo: mode == config.TradingModeDemo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
