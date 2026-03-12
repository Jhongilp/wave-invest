package etoro

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
		userKey:    os.Getenv("ETORO_API_SECRET"),
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

// InstrumentsResponse represents the response from the instruments metadata endpoint
type InstrumentsResponse struct {
	InstrumentDisplayDatas []InstrumentDisplayData `json:"instrumentDisplayDatas"`
}

// InstrumentDisplayData represents instrument metadata
type InstrumentDisplayData struct {
	InstrumentID          int    `json:"instrumentID"`
	InstrumentDisplayName string `json:"instrumentDisplayName"`
	InstrumentTypeID      int    `json:"instrumentTypeID"`
	ExchangeID            int    `json:"exchangeID"`
	SymbolFull            string `json:"symbolFull"`
	StocksIndustryID      int    `json:"stocksIndustryId"`
	PriceSource           string `json:"priceSource"`
	HasExpirationDate     bool   `json:"hasExpirationDate"`
	IsInternalInstrument  bool   `json:"isInternalInstrument"`
}

// doRequest performs an HTTP request with standard eToro headers
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("x-request-id", uuid.New().String())
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("x-user-key", c.userKey)
	req.Header.Set("Accept", "application/json")

	log.Printf("Request URL: %s", req.URL.String())
	if len(c.apiKey) > 10 {
		log.Printf("API key starts with: %s...", c.apiKey[:10])
	}
	if len(c.userKey) > 10 {
		log.Printf("User key starts with: %s...", c.userKey[:10])
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eToro API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetInstrumentMetadata fetches metadata for a list of instrument IDs
func (c *Client) GetInstrumentMetadata(instrumentIDs []int) (map[int]InstrumentDisplayData, error) {
	if len(instrumentIDs) == 0 {
		return make(map[int]InstrumentDisplayData), nil
	}

	result := make(map[int]InstrumentDisplayData)
	batchSize := 50 // Batch size for API requests

	for i := 0; i < len(instrumentIDs); i += batchSize {
		end := i + batchSize
		if end > len(instrumentIDs) {
			end = len(instrumentIDs)
		}
		batch := instrumentIDs[i:end]

		// Convert IDs to comma-separated string
		idStrs := make([]string, len(batch))
		for j, id := range batch {
			idStrs[j] = strconv.Itoa(id)
		}

		req, err := http.NewRequest("GET", c.baseURL+"/api/v1/market-data/instruments", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set raw query without URL-encoding commas
		req.URL.RawQuery = "instrumentIds=" + strings.Join(idStrs, ",")

		log.Printf("Fetching instrument metadata batch %d-%d of %d", i, end, len(instrumentIDs))

		body, err := c.doRequest(req)
		if err != nil {
			log.Printf("Error fetching instrument metadata: %v", err)
			continue // Skip this batch but continue with others
		}

		var response InstrumentsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			log.Printf("Error unmarshaling instrument response: %v", err)
			continue
		}

		for _, inst := range response.InstrumentDisplayDatas {
			result[inst.InstrumentID] = inst
		}
	}

	log.Printf("Fetched metadata for %d instruments", len(result))
	return result, nil
}

// GetWatchlist fetches the user's default watchlist from eToro
func (c *Client) GetWatchlist() ([]models.Ticker, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/v1/watchlists/default-watchlists/items", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("itemsPerPage", "100")
	req.URL.RawQuery = q.Encode()

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var items []WatchlistItemDto
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("eToro returned %d watchlist items", len(items))

	// Collect instrument IDs
	instrumentIDs := make([]int, 0)
	for _, item := range items {
		if item.ItemType == "Instrument" {
			instrumentIDs = append(instrumentIDs, item.ItemID)
		}
	}

	// Fetch instrument metadata
	instrumentMap, err := c.GetInstrumentMetadata(instrumentIDs)
	if err != nil {
		log.Printf("Warning: failed to fetch instrument metadata: %v", err)
		// Continue with IDs as symbols
	}

	// Convert to Ticker models
	tickers := make([]models.Ticker, 0, len(instrumentIDs))
	for _, item := range items {
		if item.ItemType == "Instrument" {
			ticker := models.Ticker{
				Symbol:        strconv.Itoa(item.ItemID),
				Name:          fmt.Sprintf("Instrument %d", item.ItemID),
				Price:         0,
				Change:        0,
				ChangePercent: 0,
			}

			// Use metadata if available
			if inst, ok := instrumentMap[item.ItemID]; ok {
				ticker.Symbol = inst.SymbolFull
				ticker.Name = inst.InstrumentDisplayName
			}

			tickers = append(tickers, ticker)
		}
	}

	log.Printf("Returning %d tickers", len(tickers))
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
