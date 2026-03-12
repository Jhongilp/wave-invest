package etoro

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"

	"wave_invest/internal/models"
)

type Client struct {
	apiKey     string
	userKey    string
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		apiKey:     os.Getenv("ETORO_API_KEY"),
		userKey:    os.Getenv("ETORO_USER_KEY"),
		baseURL:    "https://public-api.etoro.com",
		httpClient: &http.Client{},
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

// WatchlistItemDto represents an item from eToro's watchlist API
type WatchlistItemDto struct {
	ItemID          int        `json:"itemId"`
	ItemType        string     `json:"itemType"`
	ItemRank        int        `json:"itemRank"`
	ItemAddedReason string     `json:"itemAddedReason"`
	ItemAddedDate   string     `json:"itemAddedDate"`
	Market          *MarketDto `json:"market,omitempty"`
}

// MarketDto represents market metadata for an instrument
type MarketDto struct {
	ID                     string     `json:"id"`
	SymbolName             string     `json:"symbolName"`
	DisplayName            string     `json:"displayName"`
	AssetTypeID            int        `json:"assetTypeId"`
	AssetTypeSubCategoryID *int       `json:"assetTypeSubCategoryId"`
	ExchangeID             int        `json:"exchangeId"`
	HasExpirationDate      bool       `json:"hasExpirationDate"`
	Avatar                 *AvatarDto `json:"avatar,omitempty"`
}

// AvatarDto represents avatar images for an instrument
type AvatarDto struct {
	Small  string     `json:"small"`
	Medium string     `json:"medium"`
	Large  string     `json:"large"`
	SVG    *SVGAvatar `json:"svg,omitempty"`
}

// SVGAvatar represents SVG avatar with colors
type SVGAvatar struct {
	URL             string `json:"url"`
	BackgroundColor string `json:"backgroundColor"`
	TextColor       string `json:"textColor"`
}

// GetWatchlist fetches the user's default watchlist from eToro
func (c *Client) GetWatchlist() ([]models.Ticker, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/watchlists/default-watchlists/items", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("itemsPerPage", "100")
	req.URL.RawQuery = q.Encode()

	// Set required headers
	req.Header.Set("x-request-id", uuid.New().String())
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("x-user-key", c.userKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch watchlist: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eToro API returned status %d: %s", resp.StatusCode, string(body))
	}

	var items []WatchlistItemDto
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert WatchlistItemDto to Ticker models
	tickers := make([]models.Ticker, 0, len(items))
	for _, item := range items {
		// Only include items that have market data (instruments)
		if item.Market != nil && item.ItemType == "Instrument" {
			ticker := models.Ticker{
				Symbol: item.Market.SymbolName,
				Name:   item.Market.DisplayName,
				// Price data would need to come from a separate price endpoint
				Price:         0,
				Change:        0,
				ChangePercent: 0,
			}
			tickers = append(tickers, ticker)
		}
	}

	return tickers, nil
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
