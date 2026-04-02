package models

import "time"

type Ticker struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
}

type WatchlistResponse struct {
	Tickers []Ticker `json:"tickers"`
}

// LivePrice represents a real-time price update from eToro WebSocket
type LivePrice struct {
	Symbol    string    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Last      float64   `json:"last"`
	Timestamp time.Time `json:"timestamp"`
}
