package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"wave_invest/internal/models"
	"wave_invest/internal/services"
)

func GetWatchlist(w http.ResponseWriter, r *http.Request) {
	log.Println("GetWatchlist handler called")
	watchlistService := services.NewWatchlistService()
	tickers, err := watchlistService.GetWatchlist()
	if err != nil {
		log.Printf("Error getting watchlist: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Returning %d tickers", len(tickers))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WatchlistResponse{Tickers: tickers})
}
