package services

import (
	"context"
	"fmt"
	"time"

	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"

	"github.com/google/uuid"
)

// Executor orchestrates trade execution based on scored opportunities
type Executor struct {
	etoroClient      *etoro.Client
	portfolioService *PortfolioService
	positionService  *PositionService
	txService        *TransactionService
	positionSizer    *PositionSizer
}

// NewExecutor creates a new Executor
func NewExecutor() *Executor {
	return &Executor{
		etoroClient:      etoro.NewClient(),
		portfolioService: NewPortfolioService(),
		positionService:  NewPositionService(),
		txService:        NewTransactionService(),
		positionSizer:    NewPositionSizer(),
	}
}

// ExecuteTrades executes trades for the given opportunities
func (e *Executor) ExecuteTrades(ctx context.Context, userID string, opportunities []models.ScoredOpportunity) (*models.ExecutionResult, error) {
	result := &models.ExecutionResult{
		Success:   true,
		Positions: []models.Position{},
		Errors:    []string{},
	}

	// 1. Get portfolio
	portfolio, err := e.portfolioService.GetPortfolio(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	// 2. Get current open positions count
	openPositions, err := e.positionService.GetOpenPositions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open positions: %w", err)
	}

	currentOpenCount := len(openPositions)

	// 3. Check daily loss limit
	dailyPnL := e.calculateDailyPnL(ctx, userID)
	maxDailyLoss := portfolio.Budget * portfolio.DailyLossLimit
	if dailyPnL < -maxDailyLoss {
		result.Success = false
		result.Errors = append(result.Errors, "daily loss limit exceeded, trading halted")
		return result, nil
	}

	// 4. Execute trades for qualifying opportunities
	for _, opp := range opportunities {
		// Check if we've hit max concurrent positions
		if currentOpenCount >= portfolio.MaxConcurrentPositions {
			result.SkippedCount++
			continue
		}

		// Check minimum score threshold
		if opp.Score < portfolio.MinScoreThreshold {
			result.SkippedCount++
			continue
		}

		// Skip non-actionable biases
		if opp.TradingPlan.Trade.Bias == "neutral" {
			result.SkippedCount++
			continue
		}

		// Calculate entry price (use middle of entry zone)
		entryZone := opp.TradingPlan.Trade.EntryZone
		entryPrice := (entryZone.Low + entryZone.High) / 2

		// Calculate position size
		quantity := e.positionSizer.CalculatePositionSizeFromPortfolio(*portfolio, entryPrice)
		if quantity <= 0 {
			result.SkippedCount++
			continue
		}

		positionValue := quantity * entryPrice

		// Validate position
		valid, reason := e.positionSizer.ValidatePosition(*portfolio, currentOpenCount, positionValue)
		if !valid {
			result.SkippedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", opp.TradingPlan.Ticker, reason))
			continue
		}

		// Get eToro instrument ID for the ticker
		instrumentID, err := e.etoroClient.GetInstrumentIDBySymbol(opp.TradingPlan.Ticker)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to get instrument ID for %s: %v", opp.TradingPlan.Ticker, err))
			result.SkippedCount++
			continue
		}

		// Execute trade on eToro
		etoroReq := etoro.OpenPositionRequest{
			InstrumentID:   instrumentID,
			Amount:         positionValue,
			IsBuy:          opp.TradingPlan.Trade.Bias == "bullish",
			Leverage:       1,
			StopLossRate:   opp.TradingPlan.Trade.StopLoss,
			TakeProfitRate: opp.TradingPlan.Trade.Targets.PT1,
		}

		etoroResp, err := e.etoroClient.OpenPosition(etoroReq)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to open position for %s: %v", opp.TradingPlan.Ticker, err))
			continue
		}

		// Create position record
		position := models.Position{
			ID:         uuid.New().String(),
			UserID:     userID,
			Ticker:     opp.TradingPlan.Ticker,
			EntryPrice: entryPrice,
			Quantity:   quantity,
			StopLoss:   opp.TradingPlan.Trade.StopLoss,
			TakeProfit: opp.TradingPlan.Trade.Targets.PT1,
			Status:     "open",
			EtoroID:    fmt.Sprintf("%d", etoroResp.OrderID),
			OpenedAt:   time.Now(),
		}

		// Save position to Firestore
		if err := e.positionService.SavePosition(ctx, position); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to save position for %s: %v", opp.TradingPlan.Ticker, err))
			continue
		}

		// Log transaction
		tx := models.Transaction{
			ID:        uuid.New().String(),
			UserID:    userID,
			Type:      "BUY",
			Ticker:    opp.TradingPlan.Ticker,
			Price:     entryPrice,
			Quantity:  quantity,
			Total:     positionValue,
			Timestamp: time.Now(),
		}
		if err := e.txService.LogTransaction(ctx, tx); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to log transaction for %s: %v", opp.TradingPlan.Ticker, err))
		}

		// Update portfolio balance
		newBalance := portfolio.AvailableBalance - positionValue
		if err := e.portfolioService.UpdateBalance(ctx, userID, newBalance); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to update balance: %v", err))
		}
		portfolio.AvailableBalance = newBalance

		result.Positions = append(result.Positions, position)
		result.ExecutedCount++
		currentOpenCount++
	}

	return result, nil
}

// ClosePosition closes a specific position
func (e *Executor) ClosePosition(ctx context.Context, userID string, positionID string) error {
	// Get position from Firestore
	positions, err := e.positionService.GetOpenPositions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	var position *models.Position
	for _, p := range positions {
		if p.ID == positionID {
			position = &p
			break
		}
	}

	if position == nil {
		return fmt.Errorf("position not found: %s", positionID)
	}

	// Close on eToro
	etoroResp, err := e.etoroClient.ClosePosition(position.EtoroID)
	if err != nil {
		return fmt.Errorf("failed to close eToro position: %w", err)
	}

	// Use P&L from eToro response (includes all fees)
	pnl := etoroResp.PnL

	// Update position in Firestore
	if err := e.positionService.ClosePosition(ctx, positionID, etoroResp.CloseRate, pnl); err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// Log transaction
	tx := models.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      "SELL",
		Ticker:    position.Ticker,
		Price:     etoroResp.CloseRate,
		Quantity:  position.Quantity,
		Total:     etoroResp.CloseRate * position.Quantity,
		PnL:       pnl,
		Timestamp: time.Now(),
	}
	if err := e.txService.LogTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}

	// Update portfolio balance
	portfolio, err := e.portfolioService.GetPortfolio(ctx, userID)
	if err == nil {
		newBalance := portfolio.AvailableBalance + (position.Quantity * etoroResp.CloseRate)
		e.portfolioService.UpdateBalance(ctx, userID, newBalance)
	}

	return nil
}

// calculateDailyPnL calculates today's realized P&L
func (e *Executor) calculateDailyPnL(ctx context.Context, userID string) float64 {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	transactions, err := e.txService.GetTransactions(ctx, userID, startOfDay)
	if err != nil {
		return 0
	}

	var totalPnL float64
	for _, tx := range transactions {
		totalPnL += tx.PnL
	}

	return totalPnL
}

// GetPortfolioSummary retrieves a summary of the portfolio
func (e *Executor) GetPortfolioSummary(ctx context.Context, userID string) (*models.PortfolioSummary, error) {
	portfolio, err := e.portfolioService.GetPortfolio(ctx, userID)
	if err != nil {
		return nil, err
	}

	positions, err := e.positionService.GetOpenPositions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate invested amount
	var investedAmount float64
	for _, p := range positions {
		investedAmount += p.EntryPrice * p.Quantity
	}

	// Get daily P&L
	dailyPnL := e.calculateDailyPnL(ctx, userID)

	// Calculate total P&L from all transactions
	allTimeTx, _ := e.txService.GetTransactions(ctx, userID, time.Time{})
	var totalPnL float64
	for _, tx := range allTimeTx {
		totalPnL += tx.PnL
	}

	return &models.PortfolioSummary{
		Budget:           portfolio.Budget,
		AvailableBalance: portfolio.AvailableBalance,
		InvestedAmount:   investedAmount,
		OpenPositions:    positions,
		DailyPnL:         dailyPnL,
		TotalPnL:         totalPnL,
		Rules: models.TradingRules{
			MaxPositionPercent:     portfolio.MaxPositionPercent,
			MaxConcurrentPositions: portfolio.MaxConcurrentPositions,
			MinScoreThreshold:      portfolio.MinScoreThreshold,
			DailyLossLimit:         portfolio.DailyLossLimit,
		},
	}, nil
}
