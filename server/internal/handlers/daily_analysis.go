package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"wave_invest/internal/services"

	"github.com/go-chi/chi/v5"
)

var (
	dailyAnalysisOrchestrator *services.DailyAnalysisOrchestrator
	dailyAnalysisOnce         sync.Once
)

func getDailyAnalysisOrchestrator() *services.DailyAnalysisOrchestrator {
	dailyAnalysisOnce.Do(func() {
		dailyAnalysisOrchestrator = services.NewDailyAnalysisOrchestrator()
	})
	return dailyAnalysisOrchestrator
}

// RunDailyAnalysis handles POST /api/daily-analysis
// Triggers a full daily analysis of the watchlist
func RunDailyAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := getDailyAnalysisOrchestrator().RunDailyAnalysis(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetOpportunities handles GET /api/opportunities
// Returns today's scored opportunities
func GetOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opportunities, err := getDailyAnalysisOrchestrator().GetTodaysOpportunities(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(opportunities)
}

// GetOpportunitiesByDate handles GET /api/opportunities/{date}
// Returns opportunities for a specific date (YYYY-MM-DD)
func GetOpportunitiesByDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	date := chi.URLParam(r, "date")

	if date == "" {
		http.Error(w, "date parameter is required", http.StatusBadRequest)
		return
	}

	opportunities, err := getDailyAnalysisOrchestrator().GetOpportunitiesByDate(ctx, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(opportunities)
}
