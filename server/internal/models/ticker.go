package models

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
