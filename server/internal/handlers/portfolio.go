package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"wave_invest/internal/models"
	"wave_invest/internal/services"
	"wave_invest/pkg/etoro"

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

// EtoroPortfolioResponse represents the eToro portfolio data with enriched info
type EtoroPortfolioResponse struct {
	Positions []EtoroPortfolioPosition `json:"positions"`
	Credit    float64                  `json:"credit"`
}

// EtoroPortfolioPosition represents a position from eToro with symbol info
type EtoroPortfolioPosition struct {
	PositionID     string  `json:"positionId"`
	OrderID        int64   `json:"orderId"`
	Symbol         string  `json:"symbol"`
	InstrumentID   int     `json:"instrumentId"`
	OpenRate       float64 `json:"openRate"`
	Amount         float64 `json:"amount"`
	Units          float64 `json:"units"`
	IsBuy          bool    `json:"isBuy"`
	Leverage       int     `json:"leverage"`
	StopLossRate   float64 `json:"stopLossRate"`
	TakeProfitRate float64 `json:"takeProfitRate"`
	OpenDateTime   string  `json:"openDateTime"`
}

// GetEtoroPortfolio handles GET /api/etoro/portfolio
func GetEtoroPortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := etoro.NewClient()

	// First, get the watchlist to populate symbol cache
	_, _ = client.GetWatchlist()

	etoroResult, err := client.GetPortfolio()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update portfolio available balance with eToro credit
	if err := getPortfolioService().UpdateBalance(ctx, defaultUserID, etoroResult.Credit); err != nil {
		// Log but don't fail the request
	}

	// Collect unique instrument IDs for metadata lookup
	instrumentIDs := make([]int, 0, len(etoroResult.Positions))
	seenIDs := make(map[int]bool)
	for _, p := range etoroResult.Positions {
		if !seenIDs[p.InstrumentID] {
			instrumentIDs = append(instrumentIDs, p.InstrumentID)
			seenIDs[p.InstrumentID] = true
		}
	}

	// Fetch instrument metadata to get symbols
	instrumentMap, _ := client.GetInstrumentMetadata(instrumentIDs)

	// Convert to response format with symbols
	result := make([]EtoroPortfolioPosition, len(etoroResult.Positions))
	for i, p := range etoroResult.Positions {
		symbol := strconv.Itoa(p.InstrumentID)
		if inst, ok := instrumentMap[p.InstrumentID]; ok {
			symbol = inst.SymbolFull
		}

		result[i] = EtoroPortfolioPosition{
			PositionID:     strconv.FormatInt(p.PositionID, 10),
			OrderID:        p.OrderID,
			Symbol:         symbol,
			InstrumentID:   p.InstrumentID,
			OpenRate:       p.OpenRate,
			Amount:         p.Amount,
			Units:          p.Units,
			IsBuy:          p.IsBuy,
			Leverage:       p.Leverage,
			StopLossRate:   p.StopLossRate,
			TakeProfitRate: p.TakeProfitRate,
			OpenDateTime:   p.OpenDateTime,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(EtoroPortfolioResponse{Positions: result, Credit: etoroResult.Credit})
}

// SyncPositionsResponse represents the result of syncing positions
type SyncPositionsResponse struct {
	SyncedCount       int               `json:"syncedCount"`
	SkippedCount      int               `json:"skippedCount"`
	OrphanedPositions []models.Position `json:"orphanedPositions,omitempty"` // Positions in Firestore but not in eToro
	Errors            []string          `json:"errors,omitempty"`
}

// SyncPositions handles POST /api/positions/sync
// It updates Firestore positions with actual data from eToro
func SyncPositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := etoro.NewClient()

	// Get eToro positions
	_, _ = client.GetWatchlist() // Populate symbol cache
	etoroResult, err := client.GetPortfolio()
	if err != nil {
		http.Error(w, "failed to fetch eToro portfolio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update portfolio available balance with eToro credit
	if err := getPortfolioService().UpdateBalance(ctx, defaultUserID, etoroResult.Credit); err != nil {
		// Log but don't fail the request
	}

	// Get instrument metadata for symbols
	instrumentIDs := make([]int, 0, len(etoroResult.Positions))
	seenIDs := make(map[int]bool)
	for _, p := range etoroResult.Positions {
		if !seenIDs[p.InstrumentID] {
			instrumentIDs = append(instrumentIDs, p.InstrumentID)
			seenIDs[p.InstrumentID] = true
		}
	}
	instrumentMap, _ := client.GetInstrumentMetadata(instrumentIDs)

	// Create map of eToro positions by symbol for easy lookup
	etoroBySymbol := make(map[string]etoro.PortfolioPosition)
	for _, p := range etoroResult.Positions {
		symbol := strconv.Itoa(p.InstrumentID)
		if inst, ok := instrumentMap[p.InstrumentID]; ok {
			symbol = inst.SymbolFull
		}
		etoroBySymbol[symbol] = p
	}

	// Get Firestore positions
	positionSvc := getPositionService()
	firestorePositions, err := positionSvc.GetOpenPositions(ctx, defaultUserID)
	if err != nil {
		http.Error(w, "failed to fetch Firestore positions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := SyncPositionsResponse{
		OrphanedPositions: []models.Position{},
	}
	for _, fsPos := range firestorePositions {
		etoroPos, found := etoroBySymbol[fsPos.Ticker]
		if !found {
			// Position exists in Firestore but not in eToro - it's orphaned (likely closed)
			result.OrphanedPositions = append(result.OrphanedPositions, fsPos)
			continue
		}

		// Check if update is needed
		needsUpdate := fsPos.EntryPrice != etoroPos.OpenRate ||
			fsPos.Quantity != etoroPos.Units ||
			fsPos.EtoroID != strconv.FormatInt(etoroPos.PositionID, 10)

		if needsUpdate {
			// Update Firestore position with eToro data
			fsPos.EntryPrice = etoroPos.OpenRate
			fsPos.Quantity = etoroPos.Units
			fsPos.EtoroID = strconv.FormatInt(etoroPos.PositionID, 10)

			if err := positionSvc.UpdatePosition(ctx, defaultUserID, &fsPos); err != nil {
				result.Errors = append(result.Errors, "failed to sync "+fsPos.Ticker+": "+err.Error())
				continue
			}
			result.SyncedCount++
		} else {
			result.SkippedCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ReconcilePositionsResponse represents the result of reconciling positions with trading history
type ReconcilePositionsResponse struct {
	ReconciledCount int               `json:"reconciledCount"`
	SkippedCount    int               `json:"skippedCount"`
	ClosedPositions []models.Position `json:"closedPositions,omitempty"`
	Errors          []string          `json:"errors,omitempty"`
	IsDemo          bool              `json:"isDemo"`
}

// ReconcilePositions handles POST /api/positions/reconcile
// It checks orphaned Firestore positions against eToro trading history and closes them
func ReconcilePositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := etoro.NewClient()

	result := ReconcilePositionsResponse{
		ClosedPositions: []models.Position{},
		IsDemo:          client.IsDemo(),
	}

	// Get eToro portfolio to find orphaned positions
	_, _ = client.GetWatchlist() // Populate symbol cache
	etoroResult, err := client.GetPortfolio()
	if err != nil {
		http.Error(w, "failed to fetch eToro portfolio: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get instrument metadata for symbols
	instrumentIDs := make([]int, 0, len(etoroResult.Positions))
	seenIDs := make(map[int]bool)
	for _, p := range etoroResult.Positions {
		if !seenIDs[p.InstrumentID] {
			instrumentIDs = append(instrumentIDs, p.InstrumentID)
			seenIDs[p.InstrumentID] = true
		}
	}
	instrumentMap, _ := client.GetInstrumentMetadata(instrumentIDs)

	// Create map of eToro positions by symbol
	etoroBySymbol := make(map[string]bool)
	for _, p := range etoroResult.Positions {
		symbol := strconv.Itoa(p.InstrumentID)
		if inst, ok := instrumentMap[p.InstrumentID]; ok {
			symbol = inst.SymbolFull
		}
		etoroBySymbol[symbol] = true
	}

	// Get Firestore open positions
	positionSvc := getPositionService()
	firestorePositions, err := positionSvc.GetOpenPositions(ctx, defaultUserID)
	if err != nil {
		http.Error(w, "failed to fetch Firestore positions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find orphaned positions (in Firestore but not in eToro)
	var orphanedPositions []models.Position
	for _, fsPos := range firestorePositions {
		if !etoroBySymbol[fsPos.Ticker] {
			orphanedPositions = append(orphanedPositions, fsPos)
		}
	}

	if len(orphanedPositions) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// For demo accounts, we can't get trading history, just mark positions as closed without PnL
	if client.IsDemo() {
		for _, pos := range orphanedPositions {
			if err := positionSvc.ClosePosition(ctx, pos.ID, 0, 0); err != nil {
				result.Errors = append(result.Errors, "failed to close "+pos.Ticker+": "+err.Error())
				continue
			}
			pos.Status = "closed"
			pos.ClosePrice = 0
			pos.RealizedPnL = 0
			result.ClosedPositions = append(result.ClosedPositions, pos)
			result.ReconciledCount++
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// For real accounts, fetch trading history to get actual close data
	// Use minDate from 30 days ago to cover recent trades
	minDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	tradingHistory, err := client.GetTradingHistory(minDate)
	if err != nil {
		// If we can't get history, still close positions but without PnL data
		result.Errors = append(result.Errors, "could not fetch trading history: "+err.Error())
		for _, pos := range orphanedPositions {
			if err := positionSvc.ClosePosition(ctx, pos.ID, 0, 0); err != nil {
				result.Errors = append(result.Errors, "failed to close "+pos.Ticker+": "+err.Error())
				continue
			}
			pos.Status = "closed"
			result.ClosedPositions = append(result.ClosedPositions, pos)
			result.ReconciledCount++
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Fetch instrument metadata for trading history items to get symbols
	historyInstrumentIDs := make([]int, 0)
	historySeenIDs := make(map[int]bool)
	for _, h := range tradingHistory {
		if !historySeenIDs[h.InstrumentID] {
			historyInstrumentIDs = append(historyInstrumentIDs, h.InstrumentID)
			historySeenIDs[h.InstrumentID] = true
		}
	}
	historyInstrumentMap, _ := client.GetInstrumentMetadata(historyInstrumentIDs)

	// Create map of trading history by symbol
	historyBySymbol := make(map[string]etoro.TradeHistoryItem)
	for _, h := range tradingHistory {
		symbol := strconv.Itoa(h.InstrumentID)
		if inst, ok := historyInstrumentMap[h.InstrumentID]; ok {
			symbol = inst.SymbolFull
		}
		// Keep the most recent trade for each symbol
		if existing, exists := historyBySymbol[symbol]; !exists || h.CloseTimestamp > existing.CloseTimestamp {
			historyBySymbol[symbol] = h
		}
	}

	// Reconcile orphaned positions with trading history
	for _, pos := range orphanedPositions {
		history, found := historyBySymbol[pos.Ticker]
		if found {
			// Parse close timestamp
			closedAt, err := time.Parse(time.RFC3339, history.CloseTimestamp)
			if err != nil {
				closedAt = time.Now()
			}

			if err := positionSvc.ClosePositionWithDetails(ctx, pos.ID, history.CloseRate, history.NetProfit, closedAt); err != nil {
				result.Errors = append(result.Errors, "failed to close "+pos.Ticker+": "+err.Error())
				continue
			}
			pos.Status = "closed"
			pos.ClosePrice = history.CloseRate
			pos.RealizedPnL = history.NetProfit
			pos.ClosedAt = closedAt
			result.ClosedPositions = append(result.ClosedPositions, pos)
			result.ReconciledCount++
		} else {
			// Not found in trading history - close without PnL data
			if err := positionSvc.ClosePosition(ctx, pos.ID, 0, 0); err != nil {
				result.Errors = append(result.Errors, "failed to close "+pos.Ticker+": "+err.Error())
				continue
			}
			pos.Status = "closed"
			result.ClosedPositions = append(result.ClosedPositions, pos)
			result.SkippedCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetClosedPositions handles GET /api/positions/closed
func GetClosedPositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	positions, err := getPositionService().GetClosedPositions(ctx, defaultUserID)
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
