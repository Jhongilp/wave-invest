package etoro

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"wave_invest/config"
)

const (
	etoroWSURL        = "wss://ws.etoro.com/ws"
	reconnectMinDelay = 1 * time.Second
	reconnectMaxDelay = 30 * time.Second
	pingInterval      = 30 * time.Second
	pongWait          = 60 * time.Second
	writeWait         = 10 * time.Second
)

// WSMessage represents an outgoing WebSocket message
type WSMessage struct {
	ID        string      `json:"id"`
	Operation string      `json:"operation,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// WSAuthData contains authentication credentials
type WSAuthData struct {
	UserKey string `json:"userKey"`
	APIKey  string `json:"apiKey"`
}

// WSSubscribeData contains subscription request data
type WSSubscribeData struct {
	Topics   []string `json:"topics"`
	Snapshot bool     `json:"snapshot"`
}

// WSResponse represents an incoming response message
type WSResponse struct {
	ID           string       `json:"id,omitempty"`
	Success      bool         `json:"success,omitempty"`
	Operation    string       `json:"operation,omitempty"`
	ErrorMessage string       `json:"errorMessage,omitempty"`
	ErrorCode    string       `json:"errorCode,omitempty"`
	Messages     []WSPriceMsg `json:"messages,omitempty"`
}

// WSPriceMsg represents a price update message
type WSPriceMsg struct {
	Topic   string `json:"topic"`
	Content string `json:"content"` // JSON-encoded price data
	ID      string `json:"id"`
	Type    string `json:"type"`
}

// PriceData represents parsed price content
type PriceData struct {
	Ask           string `json:"Ask"`
	Bid           string `json:"Bid"`
	LastExecution string `json:"LastExecution"`
	Date          string `json:"Date"`
	PriceRateID   string `json:"PriceRateID"`
}

// LivePrice represents a processed live price update
type LivePrice struct {
	InstrumentID int       `json:"instrumentId"`
	Symbol       string    `json:"symbol"`
	Bid          float64   `json:"bid"`
	Ask          float64   `json:"ask"`
	Last         float64   `json:"last"`
	Timestamp    time.Time `json:"timestamp"`
}

// PriceHandler is called when a price update is received
type PriceHandler func(price LivePrice)

// WSClient manages the eToro WebSocket connection
type WSClient struct {
	conn          *websocket.Conn
	apiKey        string
	userKey       string
	isDemo        bool
	symbolToID    map[string]int
	idToSymbol    map[int]string
	subscriptions map[int]bool // instrument IDs currently subscribed
	handlers      []PriceHandler
	mu            sync.RWMutex
	done          chan struct{}
	reconnecting  bool
	authenticated bool
	restClient    *Client // Reference to REST client for symbol mapping
}

var (
	wsClientInstance *WSClient
	wsClientMu       sync.Mutex
)

// NewWSClient creates a new WebSocket client (singleton)
func NewWSClient() *WSClient {
	wsClientMu.Lock()
	defer wsClientMu.Unlock()

	if wsClientInstance != nil {
		return wsClientInstance
	}

	cfg := config.Get()
	if cfg == nil {
		log.Println("eToro WS: Config not loaded yet")
		return nil
	}

	wsClientInstance = &WSClient{
		apiKey:        cfg.EtoroAPIKey,
		userKey:       cfg.EtoroAPISecret,
		isDemo:        cfg.IsDemo(),
		symbolToID:    make(map[string]int),
		idToSymbol:    make(map[int]string),
		subscriptions: make(map[int]bool),
		handlers:      make([]PriceHandler, 0),
		done:          make(chan struct{}),
		restClient:    NewClient(),
	}

	return wsClientInstance
}

// Connect establishes the WebSocket connection
func (c *WSClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	// Reinitialize done channel (may have been closed on previous disconnect)
	select {
	case <-c.done:
		// Channel was closed, create new one
		c.done = make(chan struct{})
	default:
		// Channel is still open
	}

	log.Printf("eToro WS: Connecting to %s (isDemo=%v)", etoroWSURL, c.isDemo)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(etoroWSURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to eToro WebSocket: %w", err)
	}

	c.conn = conn
	c.authenticated = false

	// Set up connection parameters
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Start message reader
	go c.readMessages()

	// Start ping ticker
	go c.pingLoop()

	// Authenticate
	if err := c.authenticate(); err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("authentication failed: %w", err)
	}

	log.Println("eToro WS: Connected and authenticated")
	return nil
}

// authenticate sends authentication request
func (c *WSClient) authenticate() error {
	msg := WSMessage{
		ID:        uuid.New().String(),
		Operation: "Authenticate",
		Data: WSAuthData{
			UserKey: c.userKey,
			APIKey:  c.apiKey,
		},
	}

	return c.sendMessage(msg)
}

// Subscribe subscribes to price updates for given symbols
func (c *WSClient) Subscribe(symbols []string) error {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	// Get instrument IDs for symbols
	instrumentIDs, err := c.getInstrumentIDs(symbols)
	if err != nil {
		return fmt.Errorf("failed to get instrument IDs: %w", err)
	}

	// Build topics
	topics := make([]string, 0, len(instrumentIDs))
	for _, id := range instrumentIDs {
		c.mu.Lock()
		if !c.subscriptions[id] {
			topics = append(topics, fmt.Sprintf("instrument:%d", id))
			c.subscriptions[id] = true
		}
		c.mu.Unlock()
	}

	if len(topics) == 0 {
		return nil // All already subscribed
	}

	msg := WSMessage{
		ID:        uuid.New().String(),
		Operation: "Subscribe",
		Data: WSSubscribeData{
			Topics:   topics,
			Snapshot: true, // Get initial snapshot
		},
	}

	log.Printf("eToro WS: Subscribing to %d instruments: %v", len(topics), topics)
	return c.sendMessage(msg)
}

// Unsubscribe removes subscriptions for given symbols
func (c *WSClient) Unsubscribe(symbols []string) error {
	if c.conn == nil {
		return nil
	}

	instrumentIDs, err := c.getInstrumentIDs(symbols)
	if err != nil {
		return err
	}

	topics := make([]string, 0, len(instrumentIDs))
	for _, id := range instrumentIDs {
		c.mu.Lock()
		if c.subscriptions[id] {
			topics = append(topics, fmt.Sprintf("instrument:%d", id))
			delete(c.subscriptions, id)
		}
		c.mu.Unlock()
	}

	if len(topics) == 0 {
		return nil
	}

	msg := WSMessage{
		ID:        uuid.New().String(),
		Operation: "Unsubscribe",
		Data: WSSubscribeData{
			Topics: topics,
		},
	}

	return c.sendMessage(msg)
}

// AddHandler registers a price update handler
func (c *WSClient) AddHandler(handler PriceHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = append(c.handlers, handler)
}

// RemoveAllHandlers clears all handlers
func (c *WSClient) RemoveAllHandlers() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers = make([]PriceHandler, 0)
}

// Close closes the WebSocket connection
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close done channel (reading from closed channel returns immediately)
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}

	// Clear subscriptions (will need to resubscribe on reconnect)
	c.subscriptions = make(map[int]bool)

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.authenticated = false
		return err
	}
	return nil
}

// sendMessage sends a JSON message over WebSocket
func (c *WSClient) sendMessage(msg WSMessage) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteJSON(msg)
}

// readMessages reads and processes incoming messages
func (c *WSClient) readMessages() {
	defer func() {
		c.mu.Lock()
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.mu.Unlock()
		c.scheduleReconnect()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("eToro WS: Read error: %v", err)
			}
			return
		}

		c.handleMessage(message)
	}
}

// handleMessage processes a single incoming message
func (c *WSClient) handleMessage(data []byte) {
	var resp WSResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Printf("eToro WS: Failed to parse message: %v", err)
		return
	}

	// Handle authentication response
	if resp.Operation == "Authenticate" {
		if resp.Success {
			c.mu.Lock()
			c.authenticated = true
			c.mu.Unlock()
			log.Println("eToro WS: Authentication successful")
		} else {
			log.Printf("eToro WS: Authentication failed: %s (%s)", resp.ErrorMessage, resp.ErrorCode)
		}
		return
	}

	// Handle subscription response
	if resp.Operation == "Subscribe" {
		if resp.Success {
			log.Println("eToro WS: Subscription successful")
		} else {
			log.Printf("eToro WS: Subscription failed: %s", resp.ErrorMessage)
		}
		return
	}

	// Handle price messages
	for _, msg := range resp.Messages {
		if msg.Type == "Trading.Instrument.Rate" {
			c.handlePriceUpdate(msg)
		}
	}
}

// handlePriceUpdate processes a price update message
func (c *WSClient) handlePriceUpdate(msg WSPriceMsg) {
	// Extract instrument ID from topic (format: "instrument:12345")
	var instrumentID int
	_, err := fmt.Sscanf(msg.Topic, "instrument:%d", &instrumentID)
	if err != nil {
		log.Printf("eToro WS: Failed to parse topic: %s", msg.Topic)
		return
	}

	// Parse price content
	var priceData PriceData
	if err := json.Unmarshal([]byte(msg.Content), &priceData); err != nil {
		log.Printf("eToro WS: Failed to parse price data: %v", err)
		return
	}

	// Discard updates without price data (only have timestamp/tick name)
	if priceData.Bid == "" && priceData.Ask == "" && priceData.LastExecution == "" {
		return
	}

	// Get symbol for instrument
	c.mu.RLock()
	symbol := c.idToSymbol[instrumentID]
	c.mu.RUnlock()

	// Parse timestamp
	timestamp, _ := time.Parse(time.RFC3339Nano, priceData.Date)

	// Parse prices
	var bid, ask, last float64
	fmt.Sscanf(priceData.Bid, "%f", &bid)
	fmt.Sscanf(priceData.Ask, "%f", &ask)
	fmt.Sscanf(priceData.LastExecution, "%f", &last)

	// Discard if all prices are zero (invalid data)
	if bid == 0 && ask == 0 && last == 0 {
		return
	}

	livePrice := LivePrice{
		InstrumentID: instrumentID,
		Symbol:       symbol,
		Bid:          bid,
		Ask:          ask,
		Last:         last,
		Timestamp:    timestamp,
	}

	// Notify handlers
	c.mu.RLock()
	handlers := make([]PriceHandler, len(c.handlers))
	copy(handlers, c.handlers)
	c.mu.RUnlock()

	if len(handlers) == 0 {
		log.Printf("eToro WS: No handlers registered for price update: %s", symbol)
	}

	for _, handler := range handlers {
		handler(livePrice)
	}
}

// pingLoop sends periodic pings to keep connection alive
func (c *WSClient) pingLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn != nil {
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("eToro WS: Ping failed: %v", err)
					return
				}
			}
		}
	}
}

// scheduleReconnect attempts to reconnect with exponential backoff
func (c *WSClient) scheduleReconnect() {
	c.mu.Lock()
	if c.reconnecting {
		c.mu.Unlock()
		return
	}
	c.reconnecting = true
	c.mu.Unlock()

	delay := reconnectMinDelay

	for {
		select {
		case <-c.done:
			return
		case <-time.After(delay):
		}

		log.Printf("eToro WS: Attempting reconnect...")
		if err := c.Connect(); err != nil {
			log.Printf("eToro WS: Reconnect failed: %v", err)
			delay = time.Duration(float64(delay) * 1.5)
			if delay > reconnectMaxDelay {
				delay = reconnectMaxDelay
			}
			continue
		}

		// Resubscribe to previously subscribed instruments
		c.mu.Lock()
		c.reconnecting = false
		subs := make([]int, 0, len(c.subscriptions))
		for id := range c.subscriptions {
			subs = append(subs, id)
		}
		// Clear subscriptions so Subscribe will actually send the request
		c.subscriptions = make(map[int]bool)
		c.mu.Unlock()

		if len(subs) > 0 {
			topics := make([]string, len(subs))
			for i, id := range subs {
				topics[i] = fmt.Sprintf("instrument:%d", id)
			}

			msg := WSMessage{
				ID:        uuid.New().String(),
				Operation: "Subscribe",
				Data: WSSubscribeData{
					Topics:   topics,
					Snapshot: true,
				},
			}
			c.sendMessage(msg)
		}

		log.Println("eToro WS: Reconnected successfully")
		return
	}
}

// getInstrumentIDs converts symbols to instrument IDs
func (c *WSClient) getInstrumentIDs(symbols []string) ([]int, error) {
	ids := make([]int, 0, len(symbols))

	// Use REST client's GetInstrumentIDBySymbol which handles:
	// - Case-insensitive lookup
	// - Auto-fetching watchlist if cache is empty
	if c.restClient != nil {
		for _, sym := range symbols {
			id, err := c.restClient.GetInstrumentIDBySymbol(sym)
			if err != nil {
				log.Printf("eToro WS: Failed to get instrument ID for %s: %v", sym, err)
				continue
			}
			ids = append(ids, id)

			// Cache in WS client too
			c.mu.Lock()
			c.symbolToID[strings.ToUpper(sym)] = id
			c.idToSymbol[id] = sym
			c.mu.Unlock()
		}
	}

	return ids, nil
}

// GetSymbol returns the symbol for an instrument ID
func (c *WSClient) GetSymbol(instrumentID int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.idToSymbol[instrumentID]
}

// IsConnected returns true if WebSocket is connected
func (c *WSClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil && c.authenticated
}
