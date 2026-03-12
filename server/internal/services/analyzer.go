package services

import (
	"sync"

	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"
	"wave_invest/pkg/gemini"
)

type AnalyzerService struct {
	etoroClient  *etoro.Client
	geminiClient *gemini.Client
	cache        map[string]*models.TradingPlan
	cacheMu      sync.RWMutex
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		etoroClient:  etoro.NewClient(),
		geminiClient: gemini.NewClient(),
		cache:        make(map[string]*models.TradingPlan),
	}
}

func (s *AnalyzerService) Analyze(ticker string) (*models.TradingPlan, error) {
	// Get ticker data from eToro
	tickerData, err := s.etoroClient.GetTickerData(ticker)
	if err != nil {
		return nil, err
	}

	// Generate trading plan using Gemini AI
	plan, err := s.geminiClient.GenerateTradingPlan(ticker, tickerData)
	if err != nil {
		return nil, err
	}

	// Cache the plan
	s.cacheMu.Lock()
	s.cache[ticker] = plan
	s.cacheMu.Unlock()

	return plan, nil
}

func (s *AnalyzerService) AnalyzeBatch(tickers []string) ([]*models.TradingPlan, error) {
	plans := make([]*models.TradingPlan, 0, len(tickers))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(tickers))

	for _, ticker := range tickers {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			plan, err := s.Analyze(t)
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
