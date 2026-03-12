package gemini

import (
	"os"
	"time"

	"wave_invest/internal/models"
	"wave_invest/pkg/etoro"
)

type Client struct {
	apiKey string
}

func NewClient() *Client {
	return &Client{
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

// GenerateTradingPlan uses Gemini AI to analyze ticker data and generate a trading plan
// TODO: Implement actual Gemini API integration
func (c *Client) GenerateTradingPlan(ticker string, data *etoro.TickerData) (*models.TradingPlan, error) {
	// Mock response for development
	// In production, this would call the Gemini API with a carefully crafted prompt
	currentPrice := 178.50
	if len(data.Prices) > 0 {
		currentPrice = data.Prices[0].Close
	}

	return &models.TradingPlan{
		Ticker:     ticker,
		AnalyzedAt: time.Now().UTC().Format(time.RFC3339),
		Technicals: models.Technicals{
			MA20:  currentPrice * 0.98,
			MA50:  currentPrice * 0.95,
			MA200: currentPrice * 0.88,
			RSI:   data.CurrentRSI,
			VolumeProfile: []models.VolumeLevel{
				{Price: currentPrice * 0.95, Volume: 25000000, Type: "high"},
				{Price: currentPrice * 1.02, Volume: 18000000, Type: "high"},
				{Price: currentPrice * 0.92, Volume: 12000000, Type: "low"},
			},
		},
		Levels: models.Levels{
			Support:    []float64{currentPrice * 0.95, currentPrice * 0.90, currentPrice * 0.85},
			Resistance: []float64{currentPrice * 1.05, currentPrice * 1.10, currentPrice * 1.15},
		},
		Trade: models.Trade{
			Bias: "bullish",
			EntryZone: models.EntryZone{
				Low:  currentPrice * 0.97,
				High: currentPrice * 0.99,
			},
			StopLoss: currentPrice * 0.93,
			Targets: models.Targets{
				PT1: currentPrice * 1.05,
				PT2: currentPrice * 1.10,
				PT3: currentPrice * 1.18,
			},
			RiskRewardRatio: 2.5,
		},
		Sentiment: models.Sentiment{
			Score:             65,
			InstitutionalFlow: "Net buying observed over the past 5 sessions",
			SmartMoneyBets: []string{
				"Large call options volume at $190 strike",
				"Institutional accumulation detected",
				"Insider buying reported last week",
			},
		},
		Summary: "Technical analysis suggests a bullish setup with price consolidating above key moving averages. RSI indicates room for upside. Consider entries in the $" + formatPrice(currentPrice*0.97) + "-$" + formatPrice(currentPrice*0.99) + " zone with a stop loss below $" + formatPrice(currentPrice*0.93) + ". Multiple price targets offer favorable risk/reward ratio of 2.5:1.",
	}, nil
}

func formatPrice(price float64) string {
	return string(rune(int(price)))
}
