package main

import (
	"log"
	"net/http"
	"os"

	"wave_invest/internal/handlers"

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

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/health", handlers.HealthCheck)

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
		r.Delete("/positions/{id}", handlers.ClosePosition)
		r.Get("/transactions", handlers.GetTransactions)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
