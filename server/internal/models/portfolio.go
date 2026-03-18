package models

import "time"

// Portfolio represents a user's trading portfolio configuration
type Portfolio struct {
	UserID                 string    `json:"userId" firestore:"userId"`
	Budget                 float64   `json:"budget" firestore:"budget"`
	AvailableBalance       float64   `json:"availableBalance" firestore:"availableBalance"`
	MaxPositionPercent     float64   `json:"maxPositionPercent" firestore:"maxPositionPercent"`         // e.g., 0.05 for 5%
	MaxConcurrentPositions int       `json:"maxConcurrentPositions" firestore:"maxConcurrentPositions"` // e.g., 5
	MinScoreThreshold      float64   `json:"minScoreThreshold" firestore:"minScoreThreshold"`           // e.g., 60
	DailyLossLimit         float64   `json:"dailyLossLimit" firestore:"dailyLossLimit"`                 // e.g., 0.10 for 10%
	CreatedAt              time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt" firestore:"updatedAt"`
}

// Position represents an open trading position
type Position struct {
	ID          string    `json:"id" firestore:"id"`
	UserID      string    `json:"userId" firestore:"userId"`
	Ticker      string    `json:"ticker" firestore:"ticker"`
	EntryPrice  float64   `json:"entryPrice" firestore:"entryPrice"`
	Quantity    float64   `json:"quantity" firestore:"quantity"`
	StopLoss    float64   `json:"stopLoss" firestore:"stopLoss"`
	TakeProfit  float64   `json:"takeProfit" firestore:"takeProfit"` // PT1 initially
	Status      string    `json:"status" firestore:"status"`         // "open", "closed"
	EtoroID     string    `json:"etoroId" firestore:"etoroId"`       // eToro position ID
	OpenedAt    time.Time `json:"openedAt" firestore:"openedAt"`
	ClosedAt    time.Time `json:"closedAt,omitempty" firestore:"closedAt,omitempty"`
	ClosePrice  float64   `json:"closePrice,omitempty" firestore:"closePrice,omitempty"`
	RealizedPnL float64   `json:"realizedPnl,omitempty" firestore:"realizedPnl,omitempty"`
}

// Transaction represents a trading transaction (buy/sell)
type Transaction struct {
	ID        string    `json:"id" firestore:"id"`
	UserID    string    `json:"userId" firestore:"userId"`
	Type      string    `json:"type" firestore:"type"` // "BUY" or "SELL"
	Ticker    string    `json:"ticker" firestore:"ticker"`
	Price     float64   `json:"price" firestore:"price"`
	Quantity  float64   `json:"quantity" firestore:"quantity"`
	Total     float64   `json:"total" firestore:"total"` // price * quantity
	PnL       float64   `json:"pnl,omitempty" firestore:"pnl,omitempty"`
	Timestamp time.Time `json:"timestamp" firestore:"timestamp"`
}

// DailyAnalysis represents a single ticker's analysis for a specific date
type DailyAnalysis struct {
	ID          string      `json:"id" firestore:"id"`
	Date        string      `json:"date" firestore:"date"` // YYYY-MM-DD format
	Ticker      string      `json:"ticker" firestore:"ticker"`
	TradingPlan TradingPlan `json:"tradingPlan" firestore:"tradingPlan"`
	Score       float64     `json:"score" firestore:"score"`
	AnalyzedAt  time.Time   `json:"analyzedAt" firestore:"analyzedAt"`
}

// ScoredOpportunity combines a TradingPlan with a computed score for ranking
type ScoredOpportunity struct {
	TradingPlan    TradingPlan    `json:"tradingPlan"`
	Score          float64        `json:"score"`
	ScoreBreakdown ScoreBreakdown `json:"scoreBreakdown"`
}

// ScoreBreakdown shows how the score was calculated
type ScoreBreakdown struct {
	RiskRewardScore float64 `json:"riskRewardScore"` // 40% weight
	SentimentScore  float64 `json:"sentimentScore"`  // 25% weight
	RSIScore        float64 `json:"rsiScore"`        // 20% weight
	BiasScore       float64 `json:"biasScore"`       // 15% weight
}

// TradingRules represents configurable trading parameters
type TradingRules struct {
	MaxPositionPercent     float64 `json:"maxPositionPercent"`     // Max % of portfolio per trade
	MaxConcurrentPositions int     `json:"maxConcurrentPositions"` // Max open positions
	MinScoreThreshold      float64 `json:"minScoreThreshold"`      // Min score to execute trade
	DailyLossLimit         float64 `json:"dailyLossLimit"`         // Stop trading if exceeded
}

// DefaultTradingRules returns sensible defaults
func DefaultTradingRules() TradingRules {
	return TradingRules{
		MaxPositionPercent:     0.05, // 5%
		MaxConcurrentPositions: 5,
		MinScoreThreshold:      60.0,
		DailyLossLimit:         0.10, // 10%
	}
}

// PortfolioSummary provides a summary view for the client
type PortfolioSummary struct {
	Budget           float64      `json:"budget"`
	AvailableBalance float64      `json:"availableBalance"`
	InvestedAmount   float64      `json:"investedAmount"`
	OpenPositions    []Position   `json:"openPositions"`
	DailyPnL         float64      `json:"dailyPnl"`
	TotalPnL         float64      `json:"totalPnl"`
	Rules            TradingRules `json:"rules"`
}

// ExecutionResult represents the outcome of trade execution
type ExecutionResult struct {
	Success       bool       `json:"success"`
	ExecutedCount int        `json:"executedCount"`
	SkippedCount  int        `json:"skippedCount"`
	Errors        []string   `json:"errors,omitempty"`
	Positions     []Position `json:"positions,omitempty"`
}
