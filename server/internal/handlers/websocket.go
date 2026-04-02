package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"wave_invest/internal/services"
	"wave_invest/pkg/etoro"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (8KB to handle large symbol lists)
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from allowed origins
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"https://wave-invest.web.app",
			"https://wave-invest.firebaseapp.com",
		}
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// ClientMessage represents an incoming message from a client
type ClientMessage struct {
	Action  string   `json:"action"`  // "subscribe", "unsubscribe"
	Symbols []string `json:"symbols"` // List of ticker symbols
}

// ServerMessage represents an outgoing message to a client
type ServerMessage struct {
	Type      string  `json:"type"`                // "price", "status", "error"
	Symbol    string  `json:"symbol,omitempty"`    // Ticker symbol
	Bid       float64 `json:"bid,omitempty"`       // Bid price
	Ask       float64 `json:"ask,omitempty"`       // Ask price
	Last      float64 `json:"last,omitempty"`      // Last execution price
	Timestamp string  `json:"timestamp,omitempty"` // ISO timestamp
	Message   string  `json:"message,omitempty"`   // Status/error message
	Connected bool    `json:"connected,omitempty"` // Connection status
}

// Client represents a WebSocket client connection
type Client struct {
	id       string
	conn     *websocket.Conn
	priceHub *services.PriceHub
	send     chan etoro.LivePrice
}

// WebSocketHandler handles WebSocket connections with a shared PriceHub
type WebSocketHandler struct {
	priceHub *services.PriceHub
}

// NewWebSocketHandler creates a new WebSocket handler with the given PriceHub
func NewWebSocketHandler(priceHub *services.PriceHub) *WebSocketHandler {
	return &WebSocketHandler{priceHub: priceHub}
}

// HandleWebSocket handles WebSocket connections for live price updates
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket: Upgrade failed: %v", err)
		return
	}

	client := &Client{
		id:       uuid.New().String(),
		conn:     conn,
		priceHub: h.priceHub,
		send:     make(chan etoro.LivePrice, 256),
	}

	log.Printf("WebSocket: Client %s connected", client.id)

	// Send initial connection status
	client.sendStatus("connected", true)

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.priceHub.Unsubscribe(c.id)
		c.conn.Close()
		log.Printf("WebSocket: Client %s disconnected", c.id)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Log ALL read errors for debugging
			log.Printf("WebSocket: Read error for client %s: %v", c.id, err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket: Unexpected close error: %v", err)
			}
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		c.handleMessage(msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case price, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send price update
			msg := ServerMessage{
				Type:      "price",
				Symbol:    price.Symbol,
				Bid:       price.Bid,
				Ask:       price.Ask,
				Last:      price.Last,
				Timestamp: price.Timestamp.Format(time.RFC3339Nano),
			}

			if err := c.conn.WriteJSON(msg); err != nil {
				log.Printf("WebSocket: Write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming client messages
func (c *Client) handleMessage(msg ClientMessage) {
	switch msg.Action {
	case "subscribe":
		if len(msg.Symbols) == 0 {
			c.sendError("No symbols specified")
			return
		}

		if err := c.priceHub.Subscribe(c.id, msg.Symbols, c.send); err != nil {
			c.sendError("Failed to subscribe: " + err.Error())
			return
		}

		c.sendStatus("subscribed to "+formatSymbols(msg.Symbols), true)

	case "unsubscribe":
		c.priceHub.Unsubscribe(c.id)
		c.sendStatus("unsubscribed", true)

	case "update":
		if len(msg.Symbols) == 0 {
			c.sendError("No symbols specified")
			return
		}

		if err := c.priceHub.UpdateSubscription(c.id, msg.Symbols); err != nil {
			c.sendError("Failed to update subscription: " + err.Error())
			return
		}

		c.sendStatus("updated subscription", true)

	default:
		c.sendError("Unknown action: " + msg.Action)
	}
}

// sendStatus sends a status message to the client
func (c *Client) sendStatus(message string, connected bool) {
	msg := ServerMessage{
		Type:      "status",
		Message:   message,
		Connected: connected,
	}
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteJSON(msg); err != nil {
		log.Printf("WebSocket: Failed to send status: %v", err)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	msg := ServerMessage{
		Type:    "error",
		Message: message,
	}
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteJSON(msg); err != nil {
		log.Printf("WebSocket: Failed to send error: %v", err)
	}
}

// formatSymbols creates a comma-separated list of symbols for display
func formatSymbols(symbols []string) string {
	if len(symbols) <= 3 {
		result := ""
		for i, s := range symbols {
			if i > 0 {
				result += ", "
			}
			result += s
		}
		return result
	}
	return symbols[0] + ", " + symbols[1] + "... (" + string(rune('0'+len(symbols))) + " total)"
}
