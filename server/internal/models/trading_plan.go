package models

type VolumeLevel struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
	Type   string  `json:"type"` // "high" or "low"
}

type Technicals struct {
	MA20          float64       `json:"ma20"`
	MA50          float64       `json:"ma50"`
	MA200         float64       `json:"ma200"`
	RSI           float64       `json:"rsi"`
	VolumeProfile []VolumeLevel `json:"volumeProfile"`
}

type Levels struct {
	Support    []float64 `json:"support"`
	Resistance []float64 `json:"resistance"`
}

type Targets struct {
	PT1 float64 `json:"pt1"`
	PT2 float64 `json:"pt2"`
	PT3 float64 `json:"pt3"`
}

type EntryZone struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

type Trade struct {
	Bias            string    `json:"bias"` // "bullish", "bearish", "neutral"
	EntryZone       EntryZone `json:"entryZone"`
	StopLoss        float64   `json:"stopLoss"`
	Targets         Targets   `json:"targets"`
	RiskRewardRatio float64   `json:"riskRewardRatio"`
}

type Sentiment struct {
	Score             int      `json:"score"` // -100 to 100
	InstitutionalFlow string   `json:"institutionalFlow"`
	SmartMoneyBets    []string `json:"smartMoneyBets"`
}

type TradingPlan struct {
	Ticker     string     `json:"ticker"`
	AnalyzedAt string     `json:"analyzedAt"`
	Technicals Technicals `json:"technicals"`
	Levels     Levels     `json:"levels"`
	Trade      Trade      `json:"trade"`
	Sentiment  Sentiment  `json:"sentiment"`
	Summary    string     `json:"summary"`
}
