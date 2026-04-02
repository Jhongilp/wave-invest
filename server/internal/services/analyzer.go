package services

import (
	"sync"

	"wave_invest/internal/models"
	"wave_invest/pkg/gemini"
)

type AnalyzerService struct {
	geminiClient *gemini.Client
	cache        map[string]*models.TradingPlan
	cacheMu      sync.RWMutex
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		geminiClient: gemini.NewClient(),
		cache:        make(map[string]*models.TradingPlan),
	}
}

func (s *AnalyzerService) Analyze(ticker string) (*models.TradingPlan, error) {
	return s.AnalyzeWithPrice(ticker, nil)
}

// AnalyzeWithPrice analyzes a ticker with real-time price data
func (s *AnalyzerService) AnalyzeWithPrice(ticker string, priceInfo *gemini.PriceInfo) (*models.TradingPlan, error) {
	// Generate trading plan using Gemini AI with price data
	plan, err := s.geminiClient.GenerateTradingPlanWithPrice(ticker, priceInfo)
	if err != nil {
		return nil, err
	}

	// Cache the plan
	s.cacheMu.Lock()
	s.cache[ticker] = plan
	s.cacheMu.Unlock()

	return plan, nil
}

// TickerPriceInfo maps a ticker to its price info
type TickerPriceInfo map[string]*gemini.PriceInfo

func (s *AnalyzerService) AnalyzeBatch(tickers []string) ([]*models.TradingPlan, error) {
	return s.AnalyzeBatchWithPrices(tickers, nil)
}

// AnalyzeBatchWithPrices analyzes multiple tickers with their real-time price data
func (s *AnalyzerService) AnalyzeBatchWithPrices(tickers []string, prices TickerPriceInfo) ([]*models.TradingPlan, error) {
	plans := make([]*models.TradingPlan, 0, len(tickers))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(tickers))

	for _, ticker := range tickers {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			// Get price info for this ticker if available
			var priceInfo *gemini.PriceInfo
			if prices != nil {
				priceInfo = prices[t]
			}

			plan, err := s.AnalyzeWithPrice(t, priceInfo)
			if err != nil {
				errChan <- err
				return
			}
			mu.Lock()
			plans = append(plans, plan)
			mu.Unlock()
		}(ticker)
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	if err := <-errChan; err != nil {
		return nil, err
	}

	return plans, nil
}

func (s *AnalyzerService) GetCachedPlan(ticker string) (*models.TradingPlan, error) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	plan, ok := s.cache[ticker]
	if !ok {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

var ErrPlanNotFound = &PlanNotFoundError{}

type PlanNotFoundError struct{}

func (e *PlanNotFoundError) Error() string {
	return "trading plan not found"
}
