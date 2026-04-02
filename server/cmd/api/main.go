package main

import (
	"log"
	"net/http"

	"wave_invest/config"
	"wave_invest/internal/handlers"
	"wave_invest/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration (must be called after loading .env)
	cfg := config.Load()
	log.Printf("Trading mode: %s", cfg.TradingMode)

	// Initialize and start PriceHub for live price streaming
	priceHub := services.NewPriceHub()
	if err := priceHub.Start(); err != nil {
		log.Printf("Warning: Failed to start PriceHub: %v", err)
		// Continue anyway - prices will connect when first client subscribes
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"https://wave-invest.web.app",
			"https://wave-invest.firebaseapp.com",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/health", handlers.HealthCheck)

	// WebSocket endpoint for live price updates
	r.Get("/ws", handlers.HandleWebSocket)

	r.Route("/api", func(r chi.Router) {
		// Phase 1: Watchlist & Analysis
		r.Get("/watchlist", handlers.GetWatchlist)
		r.Post("/analyze/{ticker}", handlers.AnalyzeTicker)
		r.Post("/analyze/batch", handlers.AnalyzeBatch)
		r.Get("/plan/{ticker}", handlers.GetTradingPlan)

		// Phase 2: Daily Analysis & Opportunities
		r.Post("/daily-analysis", handlers.RunDailyAnalysis)
		r.Get("/opportunities", handlers.GetOpportunities)
		r.Get("/opportunities/{date}", handlers.GetOpportunitiesByDate)

		// Phase 2: Portfolio & Settings
		r.Get("/portfolio", handlers.GetPortfolio)
		r.Post("/portfolio", handlers.CreatePortfolio)
		r.Get("/settings", handlers.GetSettings)
		r.Put("/settings", handlers.UpdateSettings)

		// Phase 2: Trading Execution
		r.Post("/execute-trades", handlers.ExecuteTrades)
		r.Get("/positions", handlers.GetPositions)
		r.Get("/positions/closed", handlers.GetClosedPositions)
		r.Delete("/positions/{id}", handlers.ClosePosition)
		r.Post("/positions/sync", handlers.SyncPositions)
		r.Post("/positions/reconcile", handlers.ReconcilePositions)
		r.Get("/transactions", handlers.GetTransactions)

		// eToro Integration
		r.Get("/etoro/portfolio", handlers.GetEtoroPortfolio)

		// Trading Mode
		r.Get("/trading-mode", handlers.GetTradingMode)
		r.Put("/trading-mode", handlers.SetTradingMode)
	})

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
