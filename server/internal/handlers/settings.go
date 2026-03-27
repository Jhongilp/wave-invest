package handlers

import (
	"encoding/json"
	"net/http"

	"wave_invest/config"
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

	cfg := config.Get()
	cfg.SetTradingMode(mode)

	response := TradingModeResponse{
		Mode:   string(mode),
		IsDemo: mode == config.TradingModeDemo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
