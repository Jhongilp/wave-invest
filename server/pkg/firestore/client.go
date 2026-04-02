package firestore

import (
	"context"
	"log"
	"os"
	"sync"

	"wave_invest/config"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var (
	client *firestore.Client
	once   sync.Once
	collMu sync.RWMutex
	isDemo bool
)

func init() {
	// Register listener to update collection names when trading mode changes
	config.ModeChangeListeners = append(config.ModeChangeListeners, func() {
		collMu.Lock()
		defer collMu.Unlock()
		cfg := config.Get()
		if cfg == nil {
			return // Config not loaded yet, skip
		}
		isDemo = cfg.IsDemo()
		log.Printf("Firestore: Switched to %s collections", map[bool]string{true: "DEMO", false: "REAL"}[isDemo])
	})
}

// GetClient returns a singleton Firestore client
func GetClient() *firestore.Client {
	once.Do(func() {
		ctx := context.Background()
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			projectID = os.Getenv("FIRESTORE_PROJECT_ID")
		}
		if projectID == "" {
			log.Fatal("GOOGLE_CLOUD_PROJECT or FIRESTORE_PROJECT_ID environment variable is required")
		}

		// Set initial mode from config
		cfg := config.Get()
		isDemo = cfg.IsDemo()
		if isDemo {
			log.Println("Firestore: Using DEMO collections (demo_ prefix)")
		} else {
			log.Println("Firestore: Using REAL collections (no prefix)")
		}

		var err error
		credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialsFile != "" {
			client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credentialsFile))
		} else {
			// Use default credentials (useful in GCP environments)
			client, err = firestore.NewClient(ctx, projectID)
		}

		if err != nil {
			log.Fatalf("Failed to create Firestore client: %v", err)
		}

		log.Printf("Firestore client initialized for project: %s", projectID)
	})

	return client
}

// Close closes the Firestore client
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// getPrefix returns the current collection prefix based on trading mode
func getPrefix() string {
	collMu.RLock()
	defer collMu.RUnlock()
	if isDemo {
		return "demo_"
	}
	return ""
}

// Collection name getters (dynamically prefixed based on current mode)
func CollectionPortfolios() string   { return getPrefix() + "portfolios" }
func CollectionPositions() string    { return getPrefix() + "positions" }
func CollectionTransactions() string { return getPrefix() + "transactions" }
func CollectionAnalysis() string     { return getPrefix() + "daily_analysis" }

// CollectionSettings returns the settings collection (NOT prefixed - global setting)
func CollectionSettings() string { return "settings" }
