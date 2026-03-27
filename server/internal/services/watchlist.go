package services

import (
	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"
)

type WatchlistService struct{}

func NewWatchlistService() *WatchlistService {
	return &WatchlistService{}
}

func (s *WatchlistService) GetWatchlist() ([]models.Ticker, error) {
	return etoro.NewClient().GetWatchlist()
}
