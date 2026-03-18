package services

import (
	"wave_invest/internal/models"
)

// PositionSizer calculates position sizes based on portfolio rules
type PositionSizer struct{}

// NewPositionSizer creates a new PositionSizer
func NewPositionSizer() *PositionSizer {
	return &PositionSizer{}
}

// CalculatePositionSize determines how many units to buy
// Returns quantity based on available balance, max position %, and entry price
func (ps *PositionSizer) CalculatePositionSize(
	availableBalance float64,
	maxPositionPercent float64,
	entryPrice float64,
) float64 {
	if entryPrice <= 0 || availableBalance <= 0 || maxPositionPercent <= 0 {
		return 0
	}

	// Calculate max position value
	maxPositionValue := availableBalance * maxPositionPercent

	// Calculate quantity
	quantity := maxPositionValue / entryPrice

	return quantity
}

// CalculatePositionSizeFromPortfolio calculates position size using portfolio settings
func (ps *PositionSizer) CalculatePositionSizeFromPortfolio(
	portfolio models.Portfolio,
	entryPrice float64,
) float64 {
	return ps.CalculatePositionSize(
		portfolio.AvailableBalance,
		portfolio.MaxPositionPercent,
		entryPrice,
	)
}

// ValidatePosition checks if a position can be opened given current portfolio state
func (ps *PositionSizer) ValidatePosition(
	portfolio models.Portfolio,
	currentOpenPositions int,
	positionValue float64,
) (bool, string) {
	// Check max concurrent positions
	if currentOpenPositions >= portfolio.MaxConcurrentPositions {
		return false, "maximum concurrent positions reached"
	}

	// Check available balance
	if positionValue > portfolio.AvailableBalance {
		return false, "insufficient available balance"
	}

	return true, ""
}
