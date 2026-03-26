package etoro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	"wave_invest/config"
	"wave_invest/internal/models"
)

var (
	clientInstance *Client
	clientOnce     sync.Once
)

type Client struct {
	apiKey     string
	userKey    string
	baseURL    string
	isDemo     bool
	httpClient *http.Client
	symbolToID map[string]int // Cache: symbol -> instrument ID
	idToSymbol map[int]string // Cache: instrument ID -> symbol
}

// NewClient returns a singleton eToro client
func NewClient() *Client {
	clientOnce.Do(func() {
		cfg := config.Get()
		clientInstance = &Client{
			apiKey:     cfg.EtoroAPIKey,
			userKey:    cfg.EtoroAPISecret,
			baseURL:    "https://public-api.etoro.com",
			isDemo:     cfg.IsDemo(),
			httpClient: &http.Client{},
			symbolToID: make(map[string]int),
			idToSymbol: make(map[int]string),
		}
	})
	return clientInstance
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

	// Convert to Ticker models and cache symbol-to-ID mappings
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

			// Use metadata if available and cache the mapping
			if inst, ok := instrumentMap[item.ItemID]; ok {
				ticker.Symbol = inst.SymbolFull
				ticker.Name = inst.InstrumentDisplayName
				// Cache both directions
				c.symbolToID[strings.ToUpper(inst.SymbolFull)] = item.ItemID
				c.idToSymbol[item.ItemID] = inst.SymbolFull
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
// Uses PascalCase field names as required by eToro API
type OpenPositionRequest struct {
	InstrumentID   int     `json:"InstrumentID"`
	Amount         float64 `json:"Amount"`
	IsBuy          bool    `json:"IsBuy"`
	Leverage       int     `json:"Leverage"`
	StopLossRate   float64 `json:"StopLossRate,omitempty"`
	TakeProfitRate float64 `json:"TakeProfitRate,omitempty"`
	IsTslEnabled   bool    `json:"IsTslEnabled,omitempty"`
	IsNoStopLoss   bool    `json:"IsNoStopLoss,omitempty"`
	IsNoTakeProfit bool    `json:"IsNoTakeProfit,omitempty"`
}

// OpenPositionResponse represents the response from opening a position
type OpenPositionResponse struct {
	OrderID  int64   `json:"orderId"`
	Token    string  `json:"token"`
	Status   string  `json:"status"`
	OpenRate float64 `json:"openRate"`
}

// etoroOrderForOpen represents the orderForOpen object in eToro's response
type etoroOrderForOpen struct {
	InstrumentID   int     `json:"instrumentID"`
	Amount         float64 `json:"amount"`
	IsBuy          bool    `json:"isBuy"`
	Leverage       int     `json:"leverage"`
	StopLossRate   float64 `json:"stopLossRate"`
	TakeProfitRate float64 `json:"takeProfitRate"`
	OrderID        int64   `json:"orderID"`
	OrderType      int     `json:"orderType"`
	StatusID       int     `json:"statusID"`
	OpenDateTime   string  `json:"openDateTime"`
}

// etoroOpenPositionResponse represents eToro's actual API response
type etoroOpenPositionResponse struct {
	OrderForOpen etoroOrderForOpen `json:"orderForOpen"`
	Token        string            `json:"token"`
}

// OpenPosition opens a new trading position on eToro using market order
func (c *Client) OpenPosition(req OpenPositionRequest) (*OpenPositionResponse, error) {
	// Set defaults for optional fields
	if req.StopLossRate == 0 {
		req.IsNoStopLoss = true
	}
	if req.TakeProfitRate == 0 {
		req.IsNoTakeProfit = true
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use correct endpoint based on demo/real mode
	endpoint := "/api/v1/trading/execution/market-open-orders/by-amount"
	if c.isDemo {
		endpoint = "/api/v1/trading/execution/demo/market-open-orders/by-amount"
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(payload))
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
		OrderID:  etoroResp.OrderForOpen.OrderID,
		Token:    etoroResp.Token,
		Status:   "opened",
		OpenRate: 0, // Rate will be determined by market at execution
	}, nil
}

// ClosePositionResponse represents the response from closing a position
type ClosePositionResponse struct {
	PositionID string  `json:"positionId"`
	Token      string  `json:"token"`
	Status     string  `json:"status"`
	CloseRate  float64 `json:"closeRate"`
	PnL        float64 `json:"pnl"`
}

// etoroClosePositionResponse represents eToro's actual API response for closing
type etoroClosePositionResponse struct {
	Token string `json:"token"`
}

// ClosePosition closes an existing position on eToro
func (c *Client) ClosePosition(positionID string) (*ClosePositionResponse, error) {
	// Use correct endpoint based on demo/real mode
	endpoint := "/api/v1/trading/execution/positions/" + positionID
	if c.isDemo {
		endpoint = "/api/v1/trading/execution/demo/positions/" + positionID
	}

	httpReq, err := http.NewRequest("DELETE", c.baseURL+endpoint, nil)
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
		Token:      etoroResp.Token,
		Status:     "closed",
		CloseRate:  0, // Will need to fetch actual close rate from position details
		PnL:        0,
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
	// Use correct endpoint based on demo/real mode
	endpoint := "/api/v1/trading/execution/positions"
	if c.isDemo {
		endpoint = "/api/v1/trading/execution/demo/positions"
	}

	httpReq, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
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

// PortfolioPosition represents a position in the eToro portfolio response
type PortfolioPosition struct {
	PositionID     int64   `json:"positionID"`
	OrderID        int64   `json:"orderID"`
	InstrumentID   int     `json:"instrumentID"`
	OpenRate       float64 `json:"openRate"`
	Amount         float64 `json:"amount"`
	Units          float64 `json:"units"`
	IsBuy          bool    `json:"isBuy"`
	Leverage       int     `json:"leverage"`
	StopLossRate   float64 `json:"stopLossRate"`
	TakeProfitRate float64 `json:"takeProfitRate"`
	OpenDateTime   string  `json:"openDateTime"`
}

// etoroPortfolioResponse represents eToro's portfolio API response
type etoroPortfolioResponse struct {
	ClientPortfolio struct {
		Positions []PortfolioPosition `json:"positions"`
		Credit    float64             `json:"credit"`
	} `json:"clientPortfolio"`
}

// EtoroPortfolioResult contains positions and account credit from eToro
type EtoroPortfolioResult struct {
	Positions []PortfolioPosition
	Credit    float64
}

// GetPortfolio fetches the complete portfolio from eToro including all positions with their actual rates
func (c *Client) GetPortfolio() (*EtoroPortfolioResult, error) {
	endpoint := "/api/v1/trading/info/portfolio"
	if c.isDemo {
		endpoint = "/api/v1/trading/info/demo/portfolio"
	}

	httpReq, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	var etoroResp etoroPortfolioResponse
	if err := json.Unmarshal(body, &etoroResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &EtoroPortfolioResult{
		Positions: etoroResp.ClientPortfolio.Positions,
		Credit:    etoroResp.ClientPortfolio.Credit,
	}, nil
}

// GetPositionByOrderID finds a position in the portfolio by its order ID
func (c *Client) GetPositionByOrderID(orderID int64) (*PortfolioPosition, error) {
	result, err := c.GetPortfolio()
	if err != nil {
		return nil, err
	}

	for _, p := range result.Positions {
		if p.OrderID == orderID {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("position with order ID %d not found", orderID)
}

// GetInstrumentIDBySymbol looks up an eToro instrument ID by ticker symbol
// Uses cached mapping from GetWatchlist - the watchlist must be fetched first
func (c *Client) GetInstrumentIDBySymbol(symbol string) (int, error) {
	// Look up in cache (uppercase for case-insensitive match)
	upperSymbol := strings.ToUpper(symbol)
	if id, ok := c.symbolToID[upperSymbol]; ok {
		return id, nil
	}

	// If not in cache, try to refresh the watchlist to populate the cache
	if len(c.symbolToID) == 0 {
		_, err := c.GetWatchlist()
		if err != nil {
			return 0, fmt.Errorf("failed to load watchlist for symbol lookup: %w", err)
		}
		// Try again after loading
		if id, ok := c.symbolToID[upperSymbol]; ok {
			return id, nil
		}
	}

	return 0, fmt.Errorf("instrument not found for symbol: %s (not in watchlist)", symbol)
}
