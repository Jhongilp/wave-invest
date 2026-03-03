package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wave_invest/server/internal/config"
	"wave_invest/server/internal/handlers"
	"wave_invest/server/internal/services"
	"wave_invest/server/internal/services/polygon"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize components
	polygonClient := polygon.NewClient(cfg.PolygonAPIKey)
	hub := handlers.NewHub()
	poller := services.NewPoller(polygonClient, hub, cfg)
	gainersHandler := handlers.NewGainersHandler(poller)

	// Start WebSocket hub
	go hub.Run()

	// Start background poller
	ctx, cancel := context.WithCancel(context.Background())
	go poller.Start(ctx)

	// Set up Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(corsMiddleware(cfg.AllowedOrigins))

	// Routes
	router.GET("/health", handlers.HealthHandler)
	router.GET("/api/gainers", gainersHandler.GetGainers)
	router.GET("/ws", hub.HandleWebSocket)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if cfg.PolygonAPIKey == "" {
			log.Println("Running in DEMO MODE - using mock data")
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop poller
	cancel()
	poller.Stop()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// corsMiddleware returns a CORS middleware handler
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
