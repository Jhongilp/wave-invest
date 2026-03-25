package firestore

import (
	"context"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var (
	client           *firestore.Client
	once             sync.Once
	collectionPrefix string
)

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

		// Set collection prefix based on trading mode
		mode := os.Getenv("TRADING_MODE")
		if mode == "real" {
			collectionPrefix = ""
			log.Println("Firestore: Using REAL collections (no prefix)")
		} else {
			collectionPrefix = "demo_"
			log.Println("Firestore: Using DEMO collections (demo_ prefix)")
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

		// Initialize collection names with prefix
		initCollections()

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

// Collection name getters (with environment prefix)
var (
	CollectionPortfolios   string
	CollectionPositions    string
	CollectionTransactions string
	CollectionAnalysis     string
)

// InitCollections must be called after GetClient to set collection names
func initCollections() {
	CollectionPortfolios = collectionPrefix + "portfolios"
	CollectionPositions = collectionPrefix + "positions"
	CollectionTransactions = collectionPrefix + "transactions"
	CollectionAnalysis = collectionPrefix + "daily_analysis"
}
