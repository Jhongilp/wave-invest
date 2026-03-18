package services

import (
	"context"
	"fmt"
	"time"

	"wave_invest/internal/models"
	fs "wave_invest/pkg/firestore"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// PortfolioService handles portfolio operations in Firestore
type PortfolioService struct {
	client *firestore.Client
}

// NewPortfolioService creates a new PortfolioService
func NewPortfolioService() *PortfolioService {
	return &PortfolioService{
		client: fs.GetClient(),
	}
}

// GetPortfolio retrieves a user's portfolio
func (s *PortfolioService) GetPortfolio(ctx context.Context, userID string) (*models.Portfolio, error) {
	doc, err := s.client.Collection(fs.CollectionPortfolios).Doc(userID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	var portfolio models.Portfolio
	if err := doc.DataTo(&portfolio); err != nil {
		return nil, fmt.Errorf("failed to parse portfolio: %w", err)
	}

	return &portfolio, nil
}

// CreatePortfolio creates a new portfolio for a user
func (s *PortfolioService) CreatePortfolio(ctx context.Context, userID string, budget float64) (*models.Portfolio, error) {
	defaults := models.DefaultTradingRules()
	portfolio := models.Portfolio{
		UserID:                 userID,
		Budget:                 budget,
		AvailableBalance:       budget,
		MaxPositionPercent:     defaults.MaxPositionPercent,
		MaxConcurrentPositions: defaults.MaxConcurrentPositions,
		MinScoreThreshold:      defaults.MinScoreThreshold,
		DailyLossLimit:         defaults.DailyLossLimit,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	_, err := s.client.Collection(fs.CollectionPortfolios).Doc(userID).Set(ctx, portfolio)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}

	return &portfolio, nil
}

// UpdateBalance updates the available balance
func (s *PortfolioService) UpdateBalance(ctx context.Context, userID string, newBalance float64) error {
	_, err := s.client.Collection(fs.CollectionPortfolios).Doc(userID).Update(ctx, []firestore.Update{
		{Path: "availableBalance", Value: newBalance},
		{Path: "updatedAt", Value: time.Now()},
	})
	return err
}

// UpdateRules updates trading rules for a portfolio
func (s *PortfolioService) UpdateRules(ctx context.Context, userID string, rules models.TradingRules) error {
	_, err := s.client.Collection(fs.CollectionPortfolios).Doc(userID).Update(ctx, []firestore.Update{
		{Path: "maxPositionPercent", Value: rules.MaxPositionPercent},
		{Path: "maxConcurrentPositions", Value: rules.MaxConcurrentPositions},
		{Path: "minScoreThreshold", Value: rules.MinScoreThreshold},
		{Path: "dailyLossLimit", Value: rules.DailyLossLimit},
		{Path: "updatedAt", Value: time.Now()},
	})
	return err
}

// PositionService handles position operations in Firestore
type PositionService struct {
	client *firestore.Client
}

// NewPositionService creates a new PositionService
func NewPositionService() *PositionService {
	return &PositionService{
		client: fs.GetClient(),
	}
}

// SavePosition saves a new position to Firestore
func (s *PositionService) SavePosition(ctx context.Context, position models.Position) error {
	_, err := s.client.Collection(fs.CollectionPositions).Doc(position.ID).Set(ctx, position)
	return err
}

// GetOpenPositions retrieves all open positions for a user
func (s *PositionService) GetOpenPositions(ctx context.Context, userID string) ([]models.Position, error) {
	iter := s.client.Collection(fs.CollectionPositions).
		Where("userId", "==", userID).
		Where("status", "==", "open").
		Documents(ctx)

	var positions []models.Position
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate positions: %w", err)
		}

		var pos models.Position
		if err := doc.DataTo(&pos); err != nil {
			return nil, fmt.Errorf("failed to parse position: %w", err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// ClosePosition marks a position as closed
func (s *PositionService) ClosePosition(ctx context.Context, positionID string, closePrice float64, pnl float64) error {
	_, err := s.client.Collection(fs.CollectionPositions).Doc(positionID).Update(ctx, []firestore.Update{
		{Path: "status", Value: "closed"},
		{Path: "closedAt", Value: time.Now()},
		{Path: "closePrice", Value: closePrice},
		{Path: "realizedPnl", Value: pnl},
	})
	return err
}

// TransactionService handles transaction logging in Firestore
type TransactionService struct {
	client *firestore.Client
}

// NewTransactionService creates a new TransactionService
func NewTransactionService() *TransactionService {
	return &TransactionService{
		client: fs.GetClient(),
	}
}

// LogTransaction saves a transaction to Firestore
func (s *TransactionService) LogTransaction(ctx context.Context, tx models.Transaction) error {
	_, err := s.client.Collection(fs.CollectionTransactions).Doc(tx.ID).Set(ctx, tx)
	return err
}

// GetTransactions retrieves transactions for a user within a date range
func (s *TransactionService) GetTransactions(ctx context.Context, userID string, since time.Time) ([]models.Transaction, error) {
	iter := s.client.Collection(fs.CollectionTransactions).
		Where("userId", "==", userID).
		Where("timestamp", ">=", since).
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx)

	var transactions []models.Transaction
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate transactions: %w", err)
		}

		var tx models.Transaction
		if err := doc.DataTo(&tx); err != nil {
			return nil, fmt.Errorf("failed to parse transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// AnalysisService handles daily analysis storage in Firestore
type AnalysisService struct {
	client *firestore.Client
}

// NewAnalysisService creates a new AnalysisService
func NewAnalysisService() *AnalysisService {
	return &AnalysisService{
		client: fs.GetClient(),
	}
}

// SaveAnalysis saves a daily analysis to Firestore
func (s *AnalysisService) SaveAnalysis(ctx context.Context, analysis models.DailyAnalysis) error {
	_, err := s.client.Collection(fs.CollectionAnalysis).Doc(analysis.ID).Set(ctx, analysis)
	return err
}

// GetAnalysisByDate retrieves all analyses for a specific date
func (s *AnalysisService) GetAnalysisByDate(ctx context.Context, date string) ([]models.DailyAnalysis, error) {
	iter := s.client.Collection(fs.CollectionAnalysis).
		Where("date", "==", date).
		OrderBy("score", firestore.Desc).
		Documents(ctx)

	var analyses []models.DailyAnalysis
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate analyses: %w", err)
		}

		var analysis models.DailyAnalysis
		if err := doc.DataTo(&analysis); err != nil {
			return nil, fmt.Errorf("failed to parse analysis: %w", err)
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// CleanupOldAnalysis deletes analyses older than the retention period
func (s *AnalysisService) CleanupOldAnalysis(ctx context.Context, retentionDays int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	cutoffDate := cutoff.Format("2006-01-02")

	iter := s.client.Collection(fs.CollectionAnalysis).
		Where("date", "<", cutoffDate).
		Documents(ctx)

	batch := s.client.Batch()
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return count, fmt.Errorf("failed to iterate old analyses: %w", err)
		}

		batch.Delete(doc.Ref)
		count++

		// Firestore batches are limited to 500 operations
		if count%500 == 0 {
			if _, err := batch.Commit(ctx); err != nil {
				return count, fmt.Errorf("failed to commit batch delete: %w", err)
			}
			batch = s.client.Batch()
		}
	}

	if count%500 != 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return count, fmt.Errorf("failed to commit final batch delete: %w", err)
		}
	}

	return count, nil
}
