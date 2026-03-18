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
	client *firestore.Client
	once   sync.Once
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

// Collection names
const (
	CollectionPortfolios   = "portfolios"
	CollectionPositions    = "positions"
	CollectionTransactions = "transactions"
	CollectionAnalysis     = "daily_analysis"
)
