package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"
	"wave_invest/pkg/gemini"

	"github.com/google/uuid"
)

// DailyAnalysisOrchestrator orchestrates the daily analysis workflow
type DailyAnalysisOrchestrator struct {
	watchlistService *WatchlistService
	analyzerService  *AnalyzerService
	scorer           *Scorer
	analysisService  *AnalysisService
	etoroClient      *etoro.Client
	priceHub         *PriceHub
}

// NewDailyAnalysisOrchestrator creates a new orchestrator
func NewDailyAnalysisOrchestrator() *DailyAnalysisOrchestrator {
	return &DailyAnalysisOrchestrator{
		watchlistService: NewWatchlistService(),
		analyzerService:  NewAnalyzerService(),
		scorer:           NewScorer(),
		analysisService:  NewAnalysisService(),
		etoroClient:      etoro.NewClient(),
		priceHub:         NewPriceHub(),
	}
}

// DailyAnalysisResult contains the result of a daily analysis run
type DailyAnalysisResult struct {
	Date          string                     `json:"date"`
	AnalyzedCount int                        `json:"analyzedCount"`
	Opportunities []models.ScoredOpportunity `json:"opportunities"`
	Errors        []string                   `json:"errors,omitempty"`
}

// RunDailyAnalysis executes the full daily analysis workflow
func (o *DailyAnalysisOrchestrator) RunDailyAnalysis(ctx context.Context) (*DailyAnalysisResult, error) {
	today := time.Now().Format("2006-01-02")
	result := &DailyAnalysisResult{
		Date:          today,
		Opportunities: []models.ScoredOpportunity{},
		Errors:        []string{},
	}

	// 1. Stop WebSocket connection during analysis to avoid conflicts
	log.Println("Pausing WebSocket connections for analysis...")
	if err := o.priceHub.Stop(); err != nil {
		log.Printf("Warning: failed to stop PriceHub: %v", err)
	}

	// Ensure we restart WebSocket after analysis completes (success or failure)
	defer func() {
		log.Println("Restarting WebSocket connections...")
		if err := o.priceHub.Start(); err != nil {
			log.Printf("Warning: failed to restart PriceHub: %v", err)
		}
	}()

	// 2. Delete existing analysis for today (replace, not append)
	deletedCount, err := o.analysisService.DeleteAnalysisByDate(ctx, today)
	if err != nil {
		log.Printf("Warning: failed to delete existing analysis: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("cleanup warning: %v", err))
	} else if deletedCount > 0 {
		log.Printf("Deleted %d existing analysis records for %s", deletedCount, today)
	}

	// 3. Fetch watchlist from eToro
	tickers, err := o.watchlistService.GetWatchlist()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch watchlist: %w", err)
	}

	if len(tickers) == 0 {
		return result, nil
	}

	// 4. Extract ticker symbols
	symbols := make([]string, len(tickers))
	for i, t := range tickers {
		symbols[i] = t.Symbol
	}

	// 5. Fetch real-time bid/ask prices from eToro
	log.Printf("Fetching live prices for %d symbols...", len(symbols))
	liveRates, err := o.etoroClient.GetLiveRatesBySymbols(symbols)
	if err != nil {
		// Log warning but continue without prices
		log.Printf("Warning: failed to fetch live rates: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("live prices unavailable: %v", err))
	}

	// 6. Convert live rates to price info map for analyzer
	priceInfoMap := make(TickerPriceInfo)
	for symbol, rate := range liveRates {
		priceInfoMap[symbol] = &gemini.PriceInfo{
			Bid:  rate.Bid,
			Ask:  rate.Ask,
			Last: rate.LastExecution,
		}
		log.Printf("  %s: Bid=%.2f, Ask=%.2f", symbol, rate.Bid, rate.Ask)
	}

	// 7. Batch analyze all tickers using Gemini AI with live prices
	planPtrs, err := o.analyzerService.AnalyzeBatchWithPrices(symbols, priceInfoMap)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// 8. Convert pointers to values for scoring
	plans := make([]models.TradingPlan, 0, len(planPtrs))
	for _, p := range planPtrs {
		if p != nil {
			plans = append(plans, *p)
		}
	}

	// 9. Score each trading plan
	opportunities := o.scorer.ScoreBatch(plans)
	result.AnalyzedCount = len(plans)
	result.Opportunities = opportunities

	// 10. Save to Firestore
	for _, opp := range opportunities {
		analysis := models.DailyAnalysis{
			ID:          uuid.New().String(),
			Date:        today,
			Ticker:      opp.TradingPlan.Ticker,
			TradingPlan: opp.TradingPlan,
			Score:       opp.Score,
			AnalyzedAt:  time.Now(),
		}
		if err := o.analysisService.SaveAnalysis(ctx, analysis); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to save analysis for %s: %v", opp.TradingPlan.Ticker, err))
		}
	}

	// 11. Cleanup old analyses (7-day retention)
	cleanupCount, err := o.analysisService.CleanupOldAnalysis(ctx, 7)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cleanup warning: %v", err))
	} else if cleanupCount > 0 {
		log.Printf("Cleaned up %d old analysis records", cleanupCount)
	}

	return result, nil
}

// GetTodaysOpportunities retrieves today's scored opportunities from Firestore
func (o *DailyAnalysisOrchestrator) GetTodaysOpportunities(ctx context.Context) ([]models.ScoredOpportunity, error) {
	today := time.Now().Format("2006-01-02")
	return o.GetOpportunitiesByDate(ctx, today)
}

// GetOpportunitiesByDate retrieves opportunities for a specific date
func (o *DailyAnalysisOrchestrator) GetOpportunitiesByDate(ctx context.Context, date string) ([]models.ScoredOpportunity, error) {
	analyses, err := o.analysisService.GetAnalysisByDate(ctx, date)
	if err != nil {
		return nil, err
	}

	opportunities := make([]models.ScoredOpportunity, len(analyses))
	for i, a := range analyses {
		// Re-score to get the breakdown (Firestore only stores the total score)
		scored := o.scorer.Score(a.TradingPlan)
		opportunities[i] = *scored
	}

	return opportunities, nil
}
