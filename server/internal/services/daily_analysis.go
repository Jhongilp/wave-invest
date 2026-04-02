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
}

// NewDailyAnalysisOrchestrator creates a new orchestrator
func NewDailyAnalysisOrchestrator() *DailyAnalysisOrchestrator {
	return &DailyAnalysisOrchestrator{
		watchlistService: NewWatchlistService(),
		analyzerService:  NewAnalyzerService(),
		scorer:           NewScorer(),
		analysisService:  NewAnalysisService(),
		etoroClient:      etoro.NewClient(),
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

	// 1. Fetch watchlist from eToro
	tickers, err := o.watchlistService.GetWatchlist()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch watchlist: %w", err)
	}

	if len(tickers) == 0 {
		return result, nil
	}

	// 2. Extract ticker symbols
	symbols := make([]string, len(tickers))
	for i, t := range tickers {
		symbols[i] = t.Symbol
	}

	// 3. Fetch real-time bid/ask prices from eToro
	log.Printf("Fetching live prices for %d symbols...", len(symbols))
	liveRates, err := o.etoroClient.GetLiveRatesBySymbols(symbols)
	if err != nil {
		// Log warning but continue without prices
		log.Printf("Warning: failed to fetch live rates: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("live prices unavailable: %v", err))
	}

	// 4. Convert live rates to price info map for analyzer
	priceInfoMap := make(TickerPriceInfo)
	for symbol, rate := range liveRates {
		priceInfoMap[symbol] = &gemini.PriceInfo{
			Bid:  rate.Bid,
			Ask:  rate.Ask,
			Last: rate.LastExecution,
		}
		log.Printf("  %s: Bid=%.2f, Ask=%.2f", symbol, rate.Bid, rate.Ask)
	}

	// 5. Batch analyze all tickers using Gemini AI with live prices
	planPtrs, err := o.analyzerService.AnalyzeBatchWithPrices(symbols, priceInfoMap)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	// 6. Convert pointers to values for scoring
	plans := make([]models.TradingPlan, 0, len(planPtrs))
	for _, p := range planPtrs {
		if p != nil {
			plans = append(plans, *p)
		}
	}

	// 7. Score each trading plan
	opportunities := o.scorer.ScoreBatch(plans)
	result.AnalyzedCount = len(plans)
	result.Opportunities = opportunities

	// 8. Save to Firestore
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

	// 9. Cleanup old analyses (7-day retention)
	cleanupCount, err := o.analysisService.CleanupOldAnalysis(ctx, 7)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cleanup warning: %v", err))
	} else if cleanupCount > 0 {
		fmt.Printf("Cleaned up %d old analysis records\n", cleanupCount)
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
