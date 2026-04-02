package services

import (
	"fmt"
	"log"
	"sync"

	"wave_invest/pkg/etoro"
)

// ClientSubscriber represents a connected client that wants price updates
type ClientSubscriber struct {
	ID      string
	Symbols []string
	Send    chan etoro.LivePrice
}

// PriceHub manages the connection to eToro WebSocket and distributes prices to clients
type PriceHub struct {
	wsClient    *etoro.WSClient
	subscribers map[string]*ClientSubscriber // client ID -> subscriber
	symbolSubs  map[string]map[string]bool   // symbol -> set of client IDs
	mu          sync.RWMutex
	running     bool
}

var (
	priceHubInstance *PriceHub
	priceHubOnce     sync.Once
)

// NewPriceHub returns the singleton PriceHub instance
func NewPriceHub() *PriceHub {
	priceHubOnce.Do(func() {
		priceHubInstance = &PriceHub{
			wsClient:    etoro.NewWSClient(),
			subscribers: make(map[string]*ClientSubscriber),
			symbolSubs:  make(map[string]map[string]bool),
		}
	})
	return priceHubInstance
}

// Start initializes the PriceHub (does NOT connect to eToro yet - connection is lazy)
func (h *PriceHub) Start() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return nil
	}
	h.running = true
	h.mu.Unlock()

	// Register price handler (connection happens lazily on first subscription)
	if h.wsClient != nil {
		h.wsClient.AddHandler(h.handlePrice)
	}

	log.Println("PriceHub: Started (will connect to eToro on first subscription)")
	return nil
}

// ensureConnected connects to eToro WebSocket if not already connected
func (h *PriceHub) ensureConnected() error {
	if h.wsClient == nil {
		return fmt.Errorf("WebSocket client not initialized")
	}
	if h.wsClient.IsConnected() {
		return nil
	}
	return h.wsClient.Connect()
}

// Stop closes the eToro WebSocket connection
func (h *PriceHub) Stop() error {
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()

	if h.wsClient != nil {
		return h.wsClient.Close()
	}
	return nil
}

// Subscribe adds a client to receive price updates for specified symbols
func (h *PriceHub) Subscribe(clientID string, symbols []string, sendChan chan etoro.LivePrice) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create or update subscriber
	subscriber := &ClientSubscriber{
		ID:      clientID,
		Symbols: symbols,
		Send:    sendChan,
	}
	h.subscribers[clientID] = subscriber

	// Track symbol subscriptions
	newSymbols := make([]string, 0)
	for _, sym := range symbols {
		if h.symbolSubs[sym] == nil {
			h.symbolSubs[sym] = make(map[string]bool)
			newSymbols = append(newSymbols, sym)
		}
		h.symbolSubs[sym][clientID] = true
	}

	// Subscribe to new symbols on eToro if needed
	if len(newSymbols) > 0 && h.wsClient != nil {
		go func() {
			// Ensure connected before subscribing
			if err := h.ensureConnected(); err != nil {
				log.Printf("PriceHub: Failed to connect to eToro: %v", err)
				return
			}
			if err := h.wsClient.Subscribe(newSymbols); err != nil {
				log.Printf("PriceHub: Failed to subscribe to eToro: %v", err)
			}
		}()
	}

	log.Printf("PriceHub: Client %s subscribed to %d symbols", clientID, len(symbols))
	return nil
}

// Unsubscribe removes a client from receiving price updates
func (h *PriceHub) Unsubscribe(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subscriber, ok := h.subscribers[clientID]
	if !ok {
		return
	}

	// Remove from symbol subscriptions
	orphanedSymbols := make([]string, 0)
	for _, sym := range subscriber.Symbols {
		if clients, ok := h.symbolSubs[sym]; ok {
			delete(clients, clientID)
			if len(clients) == 0 {
				delete(h.symbolSubs, sym)
				orphanedSymbols = append(orphanedSymbols, sym)
			}
		}
	}

	// Close send channel
	close(subscriber.Send)
	delete(h.subscribers, clientID)

	// Unsubscribe from orphaned symbols on eToro
	if len(orphanedSymbols) > 0 && h.wsClient != nil {
		go func() {
			if err := h.wsClient.Unsubscribe(orphanedSymbols); err != nil {
				log.Printf("PriceHub: Failed to unsubscribe from eToro: %v", err)
			}
		}()
	}

	log.Printf("PriceHub: Client %s unsubscribed", clientID)
}

// UpdateSubscription updates a client's subscribed symbols
func (h *PriceHub) UpdateSubscription(clientID string, symbols []string) error {
	h.mu.Lock()
	subscriber, ok := h.subscribers[clientID]
	sendChan := subscriber.Send
	h.mu.Unlock()

	if !ok {
		return nil
	}

	// First unsubscribe (without closing channel)
	h.mu.Lock()
	orphanedSymbols := make([]string, 0)
	for _, sym := range subscriber.Symbols {
		if clients, ok := h.symbolSubs[sym]; ok {
			delete(clients, clientID)
			if len(clients) == 0 {
				delete(h.symbolSubs, sym)
				orphanedSymbols = append(orphanedSymbols, sym)
			}
		}
	}
	h.mu.Unlock()

	// Then subscribe to new symbols
	h.mu.Lock()
	subscriber.Symbols = symbols
	newSymbols := make([]string, 0)
	for _, sym := range symbols {
		if h.symbolSubs[sym] == nil {
			h.symbolSubs[sym] = make(map[string]bool)
			newSymbols = append(newSymbols, sym)
		}
		h.symbolSubs[sym][clientID] = true
	}
	h.mu.Unlock()

	// Update eToro subscriptions
	if h.wsClient != nil {
		if len(orphanedSymbols) > 0 {
			go h.wsClient.Unsubscribe(orphanedSymbols)
		}
		if len(newSymbols) > 0 {
			go h.wsClient.Subscribe(newSymbols)
		}
	}

	log.Printf("PriceHub: Client %s updated subscription to %d symbols", clientID, len(symbols))
	_ = sendChan // Keep the existing channel
	return nil
}

// handlePrice is called when a price update is received from eToro
func (h *PriceHub) handlePrice(price etoro.LivePrice) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Find clients subscribed to this symbol
	clients, ok := h.symbolSubs[price.Symbol]
	if !ok {
		return
	}

	// Send to all subscribed clients (non-blocking)
	for clientID := range clients {
		if subscriber, ok := h.subscribers[clientID]; ok {
			select {
			case subscriber.Send <- price:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetSubscribedSymbols returns all currently subscribed symbols
func (h *PriceHub) GetSubscribedSymbols() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	symbols := make([]string, 0, len(h.symbolSubs))
	for sym := range h.symbolSubs {
		symbols = append(symbols, sym)
	}
	return symbols
}

// GetClientCount returns the number of connected clients
func (h *PriceHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subscribers)
}

// IsConnected returns true if the eToro WebSocket is connected
func (h *PriceHub) IsConnected() bool {
	if h.wsClient == nil {
		return false
	}
	return h.wsClient.IsConnected()
}
