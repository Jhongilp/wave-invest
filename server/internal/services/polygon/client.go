package polygon

import (
	"context"
	"log"
	"sort"
	"time"

	market "wave_invest/server/internal/models"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
)

type Client struct {
	client *polygon.Client
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		client: polygon.New(apiKey),
		apiKey: apiKey,
	}
}

// FetchAllSnapshots retrieves all stock ticker snapshots from Polygon
func (c *Client) FetchAllSnapshots(ctx context.Context) ([]market.TickerSnapshot, error) {
	params := models.GetAllTickersSnapshotParams{
		Locale:     models.US,
		MarketType: models.Stocks,
	}

	resp, err := c.client.GetAllTickersSnapshot(ctx, &params)
	if err != nil {
		return nil, err
	}

	snapshots := make([]market.TickerSnapshot, 0, len(resp.Tickers))
	for _, t := range resp.Tickers {
		snapshot := market.TickerSnapshot{
			Ticker:           t.Ticker,
			TodaysChange:     t.TodaysChange,
			TodaysChangePerc: t.TodaysChangePerc,
			Updated:          time.Time(t.Updated).UnixNano(),
		}

		// Day data (value type, not pointer)
		snapshot.Day = market.DayData{
			Open:   t.Day.Open,
			High:   t.Day.High,
			Low:    t.Day.Low,
			Close:  t.Day.Close,
			Volume: t.Day.Volume,
			VWAP:   0, // VWAP not available in snapshot
		}

		// Previous day data (value type, not pointer)
		snapshot.PrevDay = market.DayData{
			Open:   t.PrevDay.Open,
			High:   t.PrevDay.High,
			Low:    t.PrevDay.Low,
			Close:  t.PrevDay.Close,
			Volume: t.PrevDay.Volume,
			VWAP:   0, // VWAP not available in snapshot
		}

		snapshots = append(snapshots, snapshot)
	}

	log.Printf("Fetched %d ticker snapshots from Polygon", len(snapshots))
	return snapshots, nil
}

// CalculateTopGainers returns the top N gainers sorted by percentage change
func (c *Client) CalculateTopGainers(snapshots []market.TickerSnapshot, limit int) []market.Gainer {
	// Filter out tickers with invalid data
	validSnapshots := make([]market.TickerSnapshot, 0, len(snapshots))
	for _, s := range snapshots {
		// Skip if no price data or zero previous close
		if s.Day.Close <= 0 || s.PrevDay.Close <= 0 {
			continue
		}
		// Skip if change percent is 0 or negative (we want gainers)
		if s.TodaysChangePerc <= 0 {
			continue
		}
		// Skip penny stocks (price < $1) and very high percentage changes (likely errors)
		if s.Day.Close < 1 || s.TodaysChangePerc > 500 {
			continue
		}
		validSnapshots = append(validSnapshots, s)
	}

	// Sort by percentage change descending
	sort.Slice(validSnapshots, func(i, j int) bool {
		return validSnapshots[i].TodaysChangePerc > validSnapshots[j].TodaysChangePerc
	})

	// Take top N
	if len(validSnapshots) > limit {
		validSnapshots = validSnapshots[:limit]
	}

	// Convert to Gainer model
	gainers := make([]market.Gainer, len(validSnapshots))
	for i, s := range validSnapshots {
		gainers[i] = market.Gainer{
			Ticker:        s.Ticker,
			Price:         s.Day.Close,
			ChangePercent: s.TodaysChangePerc,
			Change:        s.TodaysChange,
			Volume:        int64(s.Day.Volume),
			PreviousClose: s.PrevDay.Close,
			Open:          s.Day.Open,
			High:          s.Day.High,
			Low:           s.Day.Low,
			VWAP:          s.Day.VWAP,
			UpdatedAt:     s.Updated,
		}
	}

	return gainers
}

// IsDemoMode returns true if running without API key
func (c *Client) IsDemoMode() bool {
	return c.apiKey == ""
}
