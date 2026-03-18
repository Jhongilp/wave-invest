package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"wave_invest/internal/models"
	"wave_invest/internal/services"

	"github.com/go-chi/chi/v5"
)

var (
	executor         *services.Executor
	executorOnce     sync.Once
	portfolioSvc     *services.PortfolioService
	portfolioSvcOnce sync.Once
	positionSvc      *services.PositionService
	positionSvcOnce  sync.Once
	txSvc            *services.TransactionService
	txSvcOnce        sync.Once
)

func getExecutor() *services.Executor {
	executorOnce.Do(func() {
		executor = services.NewExecutor()
	})
	return executor
}

func getPortfolioService() *services.PortfolioService {
	portfolioSvcOnce.Do(func() {
		portfolioSvc = services.NewPortfolioService()
	})
	return portfolioSvc
}

func getPositionService() *services.PositionService {
	positionSvcOnce.Do(func() {
		positionSvc = services.NewPositionService()
	})
	return positionSvc
}

func getTransactionService() *services.TransactionService {
	txSvcOnce.Do(func() {
		txSvc = services.NewTransactionService()
	})
	return txSvc
}

// Default user ID for single-user mode
const defaultUserID = "default"

// GetPortfolio handles GET /api/portfolio
func GetPortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary, err := getExecutor().GetPortfolioSummary(ctx, defaultUserID)
	if err != nil {
		// If portfolio doesn't exist, return empty summary
		summary = &models.PortfolioSummary{
			Rules: models.DefaultTradingRules(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// CreatePortfolioRequest represents the request to create a portfolio
type CreatePortfolioRequest struct {
	Budget float64 `json:"budget"`
}

// CreatePortfolio handles POST /api/portfolio
func CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Budget <= 0 {
		http.Error(w, "budget must be positive", http.StatusBadRequest)
		return
	}

	portfolio, err := getPortfolioService().CreatePortfolio(ctx, defaultUserID, req.Budget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(portfolio)
}

// GetSettings handles GET /api/settings
func GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	portfolio, err := getPortfolioService().GetPortfolio(ctx, defaultUserID)
	if err != nil {
		// Return defaults if no portfolio exists
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.DefaultTradingRules())
		return
	}

	rules := models.TradingRules{
		MaxPositionPercent:     portfolio.MaxPositionPercent,
		MaxConcurrentPositions: portfolio.MaxConcurrentPositions,
		MinScoreThreshold:      portfolio.MinScoreThreshold,
		DailyLossLimit:         portfolio.DailyLossLimit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// UpdateSettings handles PUT /api/settings
func UpdateSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var rules models.TradingRules
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate rules
	if rules.MaxPositionPercent <= 0 || rules.MaxPositionPercent > 1 {
		http.Error(w, "maxPositionPercent must be between 0 and 1", http.StatusBadRequest)
		return
	}
	if rules.MaxConcurrentPositions <= 0 {
		http.Error(w, "maxConcurrentPositions must be positive", http.StatusBadRequest)
		return
	}
	if rules.MinScoreThreshold < 0 || rules.MinScoreThreshold > 100 {
		http.Error(w, "minScoreThreshold must be between 0 and 100", http.StatusBadRequest)
		return
	}
	if rules.DailyLossLimit <= 0 || rules.DailyLossLimit > 1 {
		http.Error(w, "dailyLossLimit must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if err := getPortfolioService().UpdateRules(ctx, defaultUserID, rules); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// ExecuteTradesRequest represents the request to execute trades
type ExecuteTradesRequest struct {
	MaxTrades int `json:"maxTrades,omitempty"` // Optional limit on trades to execute
}

// ExecuteTrades handles POST /api/execute-trades
func ExecuteTrades(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get today's opportunities
	opportunities, err := getDailyAnalysisOrchestrator().GetTodaysOpportunities(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(opportunities) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.ExecutionResult{
			Success:       true,
			ExecutedCount: 0,
			SkippedCount:  0,
		})
		return
	}

	// Execute trades
	result, err := getExecutor().ExecuteTrades(ctx, defaultUserID, opportunities)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetPositions handles GET /api/positions
func GetPositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	positions, err := getPositionService().GetOpenPositions(ctx, defaultUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if positions == nil {
		positions = []models.Position{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

// ClosePosition handles DELETE /api/positions/{id}
func ClosePosition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	positionID := chi.URLParam(r, "id")

	if positionID == "" {
		http.Error(w, "position id is required", http.StatusBadRequest)
		return
	}

	if err := getExecutor().ClosePosition(ctx, defaultUserID, positionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTransactions handles GET /api/transactions
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	txService := services.NewTransactionService()

	// Get transactions from the last 7 days
	since := r.URL.Query().Get("since")
	var sinceTime = services.ParseTimeOrDefault(since, 7)

	transactions, err := txService.GetTransactions(ctx, defaultUserID, sinceTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
