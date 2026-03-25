package etoro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"wave_invest/config"
	"wave_invest/internal/models"
)

type Client struct {
	apiKey     string
	userKey    string
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		apiKey:     cfg.EtoroAPIKey,
		userKey:    cfg.EtoroAPISecret,
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
	ItemID          int    `json:"itemId"`
	ItemType        string `json:"itemType"`
	ItemRank        int    `json:"itemRank"`
	ItemAddedReason string `json:"itemAddedReason"`
	ItemAddedDate   string `json:"itemAddedDate"`
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

		body, err := c.doRequest(req)
		if err != nil {
			continue // Skip this batch but continue with others
		}

		var response InstrumentsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			continue
		}

		for _, inst := range response.InstrumentDisplayDatas {
			result[inst.InstrumentID] = inst
		}
	}

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

	// Collect instrument IDs
	instrumentIDs := make([]int, 0)
	for _, item := range items {
		if item.ItemType == "Instrument" {
			instrumentIDs = append(instrumentIDs, item.ItemID)
		}
	}

	// Fetch instrument metadata
	instrumentMap, _ := c.GetInstrumentMetadata(instrumentIDs)

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

// OpenPositionRequest represents a request to open a new position
type OpenPositionRequest struct {
	InstrumentID   int     `json:"instrumentId"`
	Amount         float64 `json:"amount"`
	IsBuy          bool    `json:"isBuy"`
	Leverage       int     `json:"leverage"`
	StopLossRate   float64 `json:"stopLossRate,omitempty"`
	TakeProfitRate float64 `json:"takeProfitRate,omitempty"`
}

// OpenPositionResponse represents the response from opening a position
type OpenPositionResponse struct {
	PositionID string  `json:"positionId"`
	Status     string  `json:"status"`
	OpenRate   float64 `json:"openRate"`
}

// etoroOpenPositionResponse represents eToro's actual API response
type etoroOpenPositionResponse struct {
	PositionID int64   `json:"positionId"`
	OpenRate   float64 `json:"openRate"`
	Amount     float64 `json:"amount"`
}

// OpenPosition opens a new trading position on eToro
func (c *Client) OpenPosition(req OpenPositionRequest) (*OpenPositionResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/api/v1/trade/positions", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	body, err := c.doRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to open position: %w", err)
	}

	var etoroResp etoroOpenPositionResponse
	if err := json.Unmarshal(body, &etoroResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &OpenPositionResponse{
		PositionID: strconv.FormatInt(etoroResp.PositionID, 10),
		Status:     "opened",
		OpenRate:   etoroResp.OpenRate,
	}, nil
}

// ClosePositionResponse represents the response from closing a position
type ClosePositionResponse struct {
	PositionID string  `json:"positionId"`
	Status     string  `json:"status"`
	CloseRate  float64 `json:"closeRate"`
	PnL        float64 `json:"pnl"`
}

// etoroClosePositionResponse represents eToro's actual API response for closing
type etoroClosePositionResponse struct {
	PositionID int64   `json:"positionId"`
	CloseRate  float64 `json:"closeRate"`
	NetProfit  float64 `json:"netProfit"`
}

// ClosePosition closes an existing position on eToro
func (c *Client) ClosePosition(positionID string) (*ClosePositionResponse, error) {
	httpReq, err := http.NewRequest("DELETE", c.baseURL+"/api/v1/trade/positions/"+positionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to close position: %w", err)
	}

	var etoroResp etoroClosePositionResponse
	if err := json.Unmarshal(body, &etoroResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ClosePositionResponse{
		PositionID: positionID,
		Status:     "closed",
		CloseRate:  etoroResp.CloseRate,
		PnL:        etoroResp.NetProfit,
	}, nil
}

// EtoroPosition represents a position from eToro's API
type EtoroPosition struct {
	PositionID   string  `json:"positionId"`
	InstrumentID int     `json:"instrumentId"`
	Amount       float64 `json:"amount"`
	OpenRate     float64 `json:"openRate"`
	CurrentRate  float64 `json:"currentRate"`
	PnL          float64 `json:"pnl"`
	IsBuy        bool    `json:"isBuy"`
}

// etoroPositionDto represents eToro's actual position data structure
type etoroPositionDto struct {
	PositionID   int64   `json:"positionId"`
	InstrumentID int     `json:"instrumentId"`
	Amount       float64 `json:"amount"`
	OpenRate     float64 `json:"openRate"`
	CurrentRate  float64 `json:"currentRate"`
	NetProfit    float64 `json:"netProfit"`
	IsBuy        bool    `json:"isBuy"`
}

// etoroPositionsResponse represents eToro's positions list response
type etoroPositionsResponse struct {
	Positions []etoroPositionDto `json:"positions"`
}

// GetOpenPositions fetches all open positions from eToro
func (c *Client) GetOpenPositions() ([]EtoroPosition, error) {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/api/v1/trade/positions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var etoroResp etoroPositionsResponse
	if err := json.Unmarshal(body, &etoroResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	positions := make([]EtoroPosition, len(etoroResp.Positions))
	for i, p := range etoroResp.Positions {
		positions[i] = EtoroPosition{
			PositionID:   strconv.FormatInt(p.PositionID, 10),
			InstrumentID: p.InstrumentID,
			Amount:       p.Amount,
			OpenRate:     p.OpenRate,
			CurrentRate:  p.CurrentRate,
			PnL:          p.NetProfit,
			IsBuy:        p.IsBuy,
		}
	}

	return positions, nil
}

// GetInstrumentIDBySymbol looks up an eToro instrument ID by ticker symbol
func (c *Client) GetInstrumentIDBySymbol(symbol string) (int, error) {
	// Search for the instrument by symbol
	httpReq, err := http.NewRequest("GET", c.baseURL+"/api/v1/market-data/instruments/search", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.URL.RawQuery = "query=" + symbol

	body, err := c.doRequest(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to search instrument: %w", err)
	}

	var response InstrumentsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find exact match by symbol
	for _, inst := range response.InstrumentDisplayDatas {
		if strings.EqualFold(inst.SymbolFull, symbol) || strings.EqualFold(inst.InstrumentDisplayName, symbol) {
			return inst.InstrumentID, nil
		}
	}

	// If no exact match, return the first result if available
	if len(response.InstrumentDisplayDatas) > 0 {
		return response.InstrumentDisplayDatas[0].InstrumentID, nil
	}

	return 0, fmt.Errorf("instrument not found for symbol: %s", symbol)
}
