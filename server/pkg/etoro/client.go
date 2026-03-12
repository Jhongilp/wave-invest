package etoro

import (
	"os"

	"wave_invest/internal/models"
)

type Client struct {
	apiKey    string
	apiSecret string
	baseURL   string
}

func NewClient() *Client {
	return &Client{
		apiKey:    os.Getenv("ETORO_API_KEY"),
		apiSecret: os.Getenv("ETORO_API_SECRET"),
		baseURL:   "https://api.etoro.com",
	}
}

// TickerData contains historical price and volume data
type TickerData struct {
	Symbol     string       `json:"symbol"`
	Prices     []PricePoint `json:"prices"`
	CurrentRSI float64      `json:"currentRsi"`
}

type PricePoint struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// GetWatchlist fetches the user's watchlist from eToro
// TODO: Implement actual eToro API integration
func (c *Client) GetWatchlist() ([]models.Ticker, error) {
	// Mock data for development
	return []models.Ticker{
		{Symbol: "AAPL", Name: "Apple Inc.", Price: 178.50, Change: 2.35, ChangePercent: 1.33},
		{Symbol: "MSFT", Name: "Microsoft Corporation", Price: 378.25, Change: -1.20, ChangePercent: -0.32},
		{Symbol: "GOOGL", Name: "Alphabet Inc.", Price: 141.80, Change: 3.45, ChangePercent: 2.49},
		{Symbol: "NVDA", Name: "NVIDIA Corporation", Price: 875.40, Change: 15.60, ChangePercent: 1.81},
		{Symbol: "TSLA", Name: "Tesla, Inc.", Price: 245.30, Change: -5.80, ChangePercent: -2.31},
		{Symbol: "META", Name: "Meta Platforms, Inc.", Price: 485.60, Change: 8.90, ChangePercent: 1.87},
		{Symbol: "AMZN", Name: "Amazon.com, Inc.", Price: 178.90, Change: 1.25, ChangePercent: 0.70},
		{Symbol: "AMD", Name: "Advanced Micro Devices", Price: 165.20, Change: -2.40, ChangePercent: -1.43},
	}, nil
}

// GetTickerData fetches historical price data for a ticker
// TODO: Implement actual eToro API integration
func (c *Client) GetTickerData(symbol string) (*TickerData, error) {
	// Mock data for development
	return &TickerData{
		Symbol: symbol,
		Prices: []PricePoint{
			{Date: "2026-03-11", Open: 175.00, High: 180.00, Low: 174.50, Close: 178.50, Volume: 45000000},
			{Date: "2026-03-10", Open: 173.50, High: 176.00, Low: 172.00, Close: 175.00, Volume: 38000000},
			{Date: "2026-03-09", Open: 171.00, High: 174.00, Low: 170.50, Close: 173.50, Volume: 42000000},
		},
		CurrentRSI: 58.5,
	}, nil
}
