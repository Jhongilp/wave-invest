package services

import (
	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"
)

type WatchlistService struct {
	etoroClient *etoro.Client
}

func NewWatchlistService() *WatchlistService {
	return &WatchlistService{
		etoroClient: etoro.NewClient(),
	}
}

func (s *WatchlistService) GetWatchlist() ([]models.Ticker, error) {
	return s.etoroClient.GetWatchlist()
}
