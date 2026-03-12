package handlers

import (
	"encoding/json"
	"net/http"

	"wave_invest/internal/models"
	"wave_invest/internal/services"
)

func GetWatchlist(w http.ResponseWriter, r *http.Request) {
	watchlistService := services.NewWatchlistService()
	tickers, err := watchlistService.GetWatchlist()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WatchlistResponse{Tickers: tickers})
}
