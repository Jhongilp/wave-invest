package handlers

import (
	"encoding/json"
	"net/http"

	"wave_invest/internal/services"

	"github.com/go-chi/chi/v5"
)

func AnalyzeTicker(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		http.Error(w, "ticker is required", http.StatusBadRequest)
		return
	}

	analyzerService := services.NewAnalyzerService()
	plan, err := analyzerService.Analyze(ticker)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

type BatchAnalyzeRequest struct {
	Tickers []string `json:"tickers"`
}

func AnalyzeBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Tickers) == 0 {
		http.Error(w, "at least one ticker is required", http.StatusBadRequest)
		return
	}

	analyzerService := services.NewAnalyzerService()
	plans, err := analyzerService.AnalyzeBatch(req.Tickers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

func GetTradingPlan(w http.ResponseWriter, r *http.Request) {
	ticker := chi.URLParam(r, "ticker")
	if ticker == "" {
		http.Error(w, "ticker is required", http.StatusBadRequest)
		return
	}

	analyzerService := services.NewAnalyzerService()
	plan, err := analyzerService.GetCachedPlan(ticker)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}
