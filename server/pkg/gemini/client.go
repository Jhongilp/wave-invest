package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"wave_invest/internal/models"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

// PriceInfo contains real-time bid/ask price data for analysis
type PriceInfo struct {
	Bid  float64
	Ask  float64
	Last float64
}

func NewClient() *Client {
	return &Client{
		apiKey: strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// GenerateTradingPlan uses Gemini AI with Google Search grounding to analyze a ticker
func (c *Client) GenerateTradingPlan(ticker string) (*models.TradingPlan, error) {
	return c.GenerateTradingPlanWithPrice(ticker, nil)
}

// GenerateTradingPlanWithPrice uses Gemini AI with real-time price data for precise analysis
func (c *Client) GenerateTradingPlanWithPrice(ticker string, priceInfo *PriceInfo) (*models.TradingPlan, error) {
	prompt := buildAnalysisPrompt(ticker, priceInfo)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"tools": []map[string]interface{}{
			{
				"googleSearch": map[string]interface{}{},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
			"responseSchema":   getResponseSchema(),
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview:generateContent?key=%s", c.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return parseAPIResponse(ticker, body)
}

func getResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"technicals": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ma20":  map[string]interface{}{"type": "number"},
					"ma50":  map[string]interface{}{"type": "number"},
					"ma200": map[string]interface{}{"type": "number"},
					"rsi":   map[string]interface{}{"type": "number"},
					"volumeProfile": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"price":  map[string]interface{}{"type": "number"},
								"volume": map[string]interface{}{"type": "number"},
								"type":   map[string]interface{}{"type": "string"},
							},
							"required": []string{"price", "volume", "type"},
						},
					},
				},
				"required": []string{"ma20", "ma50", "ma200", "rsi", "volumeProfile"},
			},
			"levels": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"support":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
					"resistance": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "number"}},
				},
				"required": []string{"support", "resistance"},
			},
			"trade": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bias": map[string]interface{}{"type": "string", "enum": []string{"bullish", "bearish", "neutral"}},
					"entryZone": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"low":  map[string]interface{}{"type": "number"},
							"high": map[string]interface{}{"type": "number"},
						},
						"required": []string{"low", "high"},
					},
					"stopLoss":        map[string]interface{}{"type": "number"},
					"riskRewardRatio": map[string]interface{}{"type": "number"},
					"targets": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"pt1": map[string]interface{}{"type": "number"},
							"pt2": map[string]interface{}{"type": "number"},
							"pt3": map[string]interface{}{"type": "number"},
						},
						"required": []string{"pt1", "pt2", "pt3"},
					},
				},
				"required": []string{"bias", "entryZone", "stopLoss", "targets", "riskRewardRatio"},
			},
			"sentiment": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"score":             map[string]interface{}{"type": "integer"},
					"institutionalFlow": map[string]interface{}{"type": "string"},
					"smartMoneyBets":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
				},
				"required": []string{"score", "institutionalFlow", "smartMoneyBets"},
			},
			"summary": map[string]interface{}{"type": "string"},
		},
		"required": []string{"technicals", "levels", "trade", "sentiment", "summary"},
	}
}

func buildAnalysisPrompt(ticker string, priceInfo *PriceInfo) string {
	priceContext := ""
	if priceInfo != nil && priceInfo.Ask > 0 {
		priceContext = fmt.Sprintf(`
REAL-TIME PRICE DATA (from broker):
- Current Ask (buy price): $%.2f
- Current Bid (sell price): $%.2f
- Last Execution: $%.2f

IMPORTANT: Use this exact Ask price ($%.2f) as the reference point for your analysis.
Calculate the Entry Zone, Stop Loss, and Targets based on this real execution price.
The Risk/Reward ratio should be calculated using the Ask price as the entry point.
`, priceInfo.Ask, priceInfo.Bid, priceInfo.Last, priceInfo.Ask)
	}

	return fmt.Sprintf(`You are an expert swing trading analyst. Use Google Search to get CURRENT real-time market data.

IMPORTANT: Search for the latest market data for ticker: %s
%s
Use Google Search to find:
- Recent price action and trends
- Current technical indicator values
- Recent news and sentiment

Then generate a comprehensive swing trading plan with:

1. TECHNICALS:
   - Moving Averages (MA20, MA50, MA200) - use actual current values
   - RSI (0-100) - use actual current value
   - Volume Profile with 2-3 key price levels showing high/low volume zones

2. KEY LEVELS:
   - 3 Support levels (based on current price action)
   - 3 Resistance levels (based on current price action)

3. TRADE SETUP:
   - Bias (bullish/bearish/neutral)
   - Entry Zone (low and high price range for optimal entry)
   - Stop Loss level
   - 3 Price Targets (PT1: conservative, PT2: moderate, PT3: aggressive)
   - Risk/Reward Ratio

4. SENTIMENT:
   - Score (-100 bearish to +100 bullish)
   - Institutional Flow description
   - Smart Money indicators (2-3 observations)

5. SUMMARY:
   - Concise actionable trading summary including the current price

Focus on swing trading timeframe (1-4 weeks holding period). All price levels must reflect CURRENT market prices.
The Stop Loss should be placed at a level that invalidates the trade thesis (below recent swing low for bullish).
`, ticker, priceContext)
}

// APIResponse represents the Gemini API response structure
type APIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// GeminiResponse represents the JSON structure from Gemini
type GeminiResponse struct {
	Technicals struct {
		MA20          float64 `json:"ma20"`
		MA50          float64 `json:"ma50"`
		MA200         float64 `json:"ma200"`
		RSI           float64 `json:"rsi"`
		VolumeProfile []struct {
			Price  float64 `json:"price"`
			Volume float64 `json:"volume"`
			Type   string  `json:"type"`
		} `json:"volumeProfile"`
	} `json:"technicals"`
	Levels struct {
		Support    []float64 `json:"support"`
		Resistance []float64 `json:"resistance"`
	} `json:"levels"`
	Trade struct {
		Bias      string `json:"bias"`
		EntryZone struct {
			Low  float64 `json:"low"`
			High float64 `json:"high"`
		} `json:"entryZone"`
		StopLoss float64 `json:"stopLoss"`
		Targets  struct {
			PT1 float64 `json:"pt1"`
			PT2 float64 `json:"pt2"`
			PT3 float64 `json:"pt3"`
		} `json:"targets"`
		RiskRewardRatio float64 `json:"riskRewardRatio"`
	} `json:"trade"`
	Sentiment struct {
		Score             int      `json:"score"`
		InstitutionalFlow string   `json:"institutionalFlow"`
		SmartMoneyBets    []string `json:"smartMoneyBets"`
	} `json:"sentiment"`
	Summary string `json:"summary"`
}

func parseAPIResponse(ticker string, body []byte) (*models.TradingPlan, error) {
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	textContent := apiResp.Candidates[0].Content.Parts[0].Text

	var geminiResp GeminiResponse
	if err := json.Unmarshal([]byte(textContent), &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal trading plan: %w", err)
	}

	// Convert to TradingPlan model
	volumeProfile := make([]models.VolumeLevel, len(geminiResp.Technicals.VolumeProfile))
	for i, v := range geminiResp.Technicals.VolumeProfile {
		volumeProfile[i] = models.VolumeLevel{
			Price:  v.Price,
			Volume: v.Volume,
			Type:   v.Type,
		}
	}

	plan := &models.TradingPlan{
		Ticker:     ticker,
		AnalyzedAt: time.Now().UTC().Format(time.RFC3339),
		Technicals: models.Technicals{
			MA20:          geminiResp.Technicals.MA20,
			MA50:          geminiResp.Technicals.MA50,
			MA200:         geminiResp.Technicals.MA200,
			RSI:           geminiResp.Technicals.RSI,
			VolumeProfile: volumeProfile,
		},
		Levels: models.Levels{
			Support:    geminiResp.Levels.Support,
			Resistance: geminiResp.Levels.Resistance,
		},
		Trade: models.Trade{
			Bias: geminiResp.Trade.Bias,
			EntryZone: models.EntryZone{
				Low:  geminiResp.Trade.EntryZone.Low,
				High: geminiResp.Trade.EntryZone.High,
			},
			StopLoss: geminiResp.Trade.StopLoss,
			Targets: models.Targets{
				PT1: geminiResp.Trade.Targets.PT1,
				PT2: geminiResp.Trade.Targets.PT2,
				PT3: geminiResp.Trade.Targets.PT3,
			},
			RiskRewardRatio: geminiResp.Trade.RiskRewardRatio,
		},
		Sentiment: models.Sentiment{
			Score:             geminiResp.Sentiment.Score,
			InstitutionalFlow: geminiResp.Sentiment.InstitutionalFlow,
			SmartMoneyBets:    geminiResp.Sentiment.SmartMoneyBets,
		},
		Summary: geminiResp.Summary,
	}

	return plan, nil
}
