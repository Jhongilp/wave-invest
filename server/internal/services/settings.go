package services

import (
	"context"
	"fmt"
	"log"
	"time"

	fs "wave_invest/pkg/firestore"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const settingsDocID = "app_settings"

// AppSettings represents global application settings stored in Firestore
type AppSettings struct {
	TradingMode string    `firestore:"tradingMode"`
	UpdatedAt   time.Time `firestore:"updatedAt"`
}

// SettingsService handles settings operations in Firestore
type SettingsService struct {
	client *firestore.Client
}

// NewSettingsService creates a new SettingsService
func NewSettingsService() *SettingsService {
	return &SettingsService{
		client: fs.GetClient(),
	}
}

// GetTradingMode retrieves the trading mode from Firestore
func (s *SettingsService) GetTradingMode(ctx context.Context) (string, error) {
	doc, err := s.client.Collection(fs.CollectionSettings()).Doc(settingsDocID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// No settings saved yet, return empty to use default
			return "", nil
		}
		return "", fmt.Errorf("failed to get settings: %w", err)
	}

	var settings AppSettings
	if err := doc.DataTo(&settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %w", err)
	}

	return settings.TradingMode, nil
}

// SetTradingMode saves the trading mode to Firestore
func (s *SettingsService) SetTradingMode(ctx context.Context, mode string) error {
	settings := AppSettings{
		TradingMode: mode,
		UpdatedAt:   time.Now(),
	}

	_, err := s.client.Collection(fs.CollectionSettings()).Doc(settingsDocID).Set(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	log.Printf("Settings: Trading mode saved to Firestore: %s", mode)
	return nil
}
