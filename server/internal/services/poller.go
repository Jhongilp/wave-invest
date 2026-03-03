package services

import (
	"context"
	"log"
	"sync"
	"time"

	"wave_invest/server/internal/config"
	"wave_invest/server/internal/handlers"
	"wave_invest/server/internal/models"
	"wave_invest/server/internal/services/polygon"
)

type Poller struct {
	client      *polygon.Client
	hub         *handlers.Hub
	config      *config.Config
	lastGainers []models.Gainer
	mu          sync.RWMutex
	stopCh      chan struct{}
}

func NewPoller(client *polygon.Client, hub *handlers.Hub, cfg *config.Config) *Poller {
	return &Poller{
		client: client,
		hub:    hub,
		config: cfg,
		stopCh: make(chan struct{}),
	}
}

// Start begins the polling loop
func (p *Poller) Start(ctx context.Context) {
	log.Printf("Starting poller with interval: %v", p.config.PollInterval)

	// Initial fetch
	p.fetchAndBroadcast(ctx)

	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Poller stopped: context cancelled")
			return
		case <-p.stopCh:
			log.Println("Poller stopped: stop signal received")
			return
		case <-ticker.C:
			p.fetchAndBroadcast(ctx)
		}
	}
}

// Stop gracefully stops the poller
func (p *Poller) Stop() {
	close(p.stopCh)
}

func (p *Poller) fetchAndBroadcast(ctx context.Context) {
	var gainers []models.Gainer

	if p.client.IsDemoMode() {
		// Use mock data in demo mode
		p.mu.RLock()
		if len(p.lastGainers) > 0 {
			// Simulate price changes on existing data
			gainers = polygon.SimulatePriceChange(p.lastGainers)
		} else {
			// Generate initial mock data
			gainers = polygon.MockGainers(20)
		}
		p.mu.RUnlock()
		log.Println("Generated mock gainer data (demo mode)")
	} else {
		// Fetch real data from Polygon
		snapshots, err := p.client.FetchAllSnapshots(ctx)
		if err != nil {
			log.Printf("Error fetching snapshots: %v", err)
			return
		}

		gainers = p.client.CalculateTopGainers(snapshots, 20)
		log.Printf("Calculated top %d gainers from %d snapshots", len(gainers), len(snapshots))
	}

	// Store last gainers
	p.mu.Lock()
	p.lastGainers = gainers
	p.mu.Unlock()

	// Broadcast to all connected clients
	msg := models.WebSocketMessage{
		Type:      "gainers",
		Data:      gainers,
		Timestamp: time.Now(),
	}

	p.hub.Broadcast(msg)
}

// GetLastGainers returns the most recent gainers data
func (p *Poller) GetLastGainers() []models.Gainer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastGainers
}
