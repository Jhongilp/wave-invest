package services

import (
	"math"
	"wave_invest/internal/models"
)

// Scorer calculates opportunity scores for trading plans
type Scorer struct{}

// NewScorer creates a new Scorer instance
func NewScorer() *Scorer {
	return &Scorer{}
}

// Scoring weights
const (
	WeightRiskReward = 0.40
	WeightSentiment  = 0.25
	WeightRSI        = 0.20
	WeightBias       = 0.15
)

// Score calculates a score for a TradingPlan and returns a ScoredOpportunity
func (s *Scorer) Score(plan models.TradingPlan) *models.ScoredOpportunity {
	breakdown := models.ScoreBreakdown{
		RiskRewardScore: s.scoreRiskReward(plan),
		SentimentScore:  s.scoreSentiment(plan),
		RSIScore:        s.scoreRSI(plan),
		BiasScore:       s.scoreBias(plan),
	}

	// Calculate weighted total score (0-100 scale)
	totalScore := breakdown.RiskRewardScore*WeightRiskReward +
		breakdown.SentimentScore*WeightSentiment +
		breakdown.RSIScore*WeightRSI +
		breakdown.BiasScore*WeightBias

	return &models.ScoredOpportunity{
		TradingPlan:    plan,
		Score:          totalScore,
		ScoreBreakdown: breakdown,
	}
}

// ScoreBatch scores multiple trading plans and returns sorted opportunities
func (s *Scorer) ScoreBatch(plans []models.TradingPlan) []models.ScoredOpportunity {
	opportunities := make([]models.ScoredOpportunity, 0, len(plans))

	for _, plan := range plans {
		scored := s.Score(plan)

		// Filter out invalid opportunities
		if s.isValidOpportunity(plan, *scored) {
			opportunities = append(opportunities, *scored)
		}
	}

	// Sort by score descending
	s.sortByScore(opportunities)

	return opportunities
}

// scoreRiskReward scores based on risk/reward ratio (0-100)
// Higher R/R = higher score, capped at 5:1
func (s *Scorer) scoreRiskReward(plan models.TradingPlan) float64 {
	rr := plan.Trade.RiskRewardRatio
	if rr <= 0 {
		return 0
	}

	// Scale: 1:1 = 20, 2:1 = 40, 3:1 = 60, 4:1 = 80, 5:1+ = 100
	score := rr * 20
	return math.Min(score, 100)
}

// scoreSentiment scores based on sentiment (-100 to 100 -> 0 to 100)
func (s *Scorer) scoreSentiment(plan models.TradingPlan) float64 {
	sentiment := float64(plan.Sentiment.Score)

	// For bullish bias, positive sentiment is good
	// For bearish bias, negative sentiment is good
	if plan.Trade.Bias == "bearish" {
		sentiment = -sentiment
	}

	// Convert -100..100 to 0..100
	return (sentiment + 100) / 2
}

// scoreRSI scores based on RSI position (0-100)
// For longs: lower RSI = higher score (oversold)
// For shorts: higher RSI = higher score (overbought)
func (s *Scorer) scoreRSI(plan models.TradingPlan) float64 {
	rsi := plan.Technicals.RSI

	if plan.Trade.Bias == "bullish" {
		// For longs, prefer oversold conditions (RSI < 30 is ideal)
		if rsi <= 30 {
			return 100
		} else if rsi <= 50 {
			return 100 - ((rsi - 30) * 2.5) // 30->100, 50->50
		} else if rsi <= 70 {
			return 50 - ((rsi - 50) * 2.5) // 50->50, 70->0
		}
		return 0
	} else if plan.Trade.Bias == "bearish" {
		// For shorts, prefer overbought conditions (RSI > 70 is ideal)
		if rsi >= 70 {
			return 100
		} else if rsi >= 50 {
			return (rsi - 50) * 5 // 50->0, 70->100
		}
		return 0
	}

	// Neutral bias - middle RSI is okay
	return 50
}

// scoreBias scores based on bias clarity (0-100)
func (s *Scorer) scoreBias(plan models.TradingPlan) float64 {
	switch plan.Trade.Bias {
	case "bullish", "bearish":
		return 100
	case "neutral":
		return 0
	default:
		return 0
	}
}

// isValidOpportunity filters out opportunities that shouldn't be traded
func (s *Scorer) isValidOpportunity(plan models.TradingPlan, scored models.ScoredOpportunity) bool {
	// Filter out neutral bias
	if plan.Trade.Bias == "neutral" {
		return false
	}

	// Filter out unfavorable RSI conditions
	rsi := plan.Technicals.RSI
	if plan.Trade.Bias == "bullish" && rsi > 70 {
		// Don't go long when overbought
		return false
	}
	if plan.Trade.Bias == "bearish" && rsi < 30 {
		// Don't go short when oversold
		return false
	}

	// Filter out very low R/R
	if plan.Trade.RiskRewardRatio < 1.0 {
		return false
	}

	return true
}

// sortByScore sorts opportunities by score descending (in-place)
func (s *Scorer) sortByScore(opportunities []models.ScoredOpportunity) {
	for i := 0; i < len(opportunities)-1; i++ {
		for j := i + 1; j < len(opportunities); j++ {
			if opportunities[j].Score > opportunities[i].Score {
				opportunities[i], opportunities[j] = opportunities[j], opportunities[i]
			}
		}
	}
}
