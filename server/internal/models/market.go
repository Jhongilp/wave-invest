package models

import "time"

// Gainer represents a stock that has gained value
type Gainer struct {
	Ticker        string  `json:"ticker"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"changePercent"`
	Change        float64 `json:"change"`
	Volume        int64   `json:"volume"`
	PreviousClose float64 `json:"previousClose"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	VWAP          float64 `json:"vwap"`
	UpdatedAt     int64   `json:"updatedAt"`
}

// WebSocketMessage is the message sent to connected clients
type WebSocketMessage struct {
	Type      string    `json:"type"`
	Data      []Gainer  `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

// TickerSnapshot represents the raw snapshot data from Polygon
type TickerSnapshot struct {
	Ticker           string  `json:"ticker"`
	TodaysChange     float64 `json:"todaysChange"`
	TodaysChangePerc float64 `json:"todaysChangePerc"`
	Updated          int64   `json:"updated"`
	Day              DayData `json:"day"`
	PrevDay          DayData `json:"prevDay"`
	Min              MinData `json:"min"`
}

// DayData represents daily trading data
type DayData struct {
	Open   float64 `json:"o"`
	High   float64 `json:"h"`
	Low    float64 `json:"l"`
	Close  float64 `json:"c"`
	Volume float64 `json:"v"`
	VWAP   float64 `json:"vw"`
}

// MinData represents minute-level data
type MinData struct {
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    float64 `json:"v"`
	VWAP      float64 `json:"vw"`
	Timestamp int64   `json:"t"`
}
