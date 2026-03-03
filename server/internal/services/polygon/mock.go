package polygon

import (
	"math/rand"
	"time"

	"wave_invest/server/internal/models"
)

// MockGainers generates fake gainer data for demo mode
func MockGainers(count int) []models.Gainer {
	tickers := []string{
		"AAPL", "MSFT", "GOOGL", "AMZN", "META", "NVDA", "TSLA", "AMD", "NFLX", "CRM",
		"ORCL", "INTC", "CSCO", "ADBE", "PYPL", "SQ", "SHOP", "SNOW", "PLTR", "UBER",
		"LYFT", "DOCU", "ZM", "TWLO", "DDOG", "NET", "CRWD", "OKTA", "MDB", "COIN",
	}

	rand.Seed(time.Now().UnixNano())

	gainers := make([]models.Gainer, count)
	for i := 0; i < count && i < len(tickers); i++ {
		prevClose := 50.0 + rand.Float64()*200
		changePercent := 1.0 + rand.Float64()*25 // 1% to 26%
		change := prevClose * changePercent / 100
		price := prevClose + change

		gainers[i] = models.Gainer{
			Ticker:        tickers[i],
			Price:         price,
			ChangePercent: changePercent,
			Change:        change,
			Volume:        int64(rand.Intn(50000000) + 1000000),
			PreviousClose: prevClose,
			Open:          prevClose + rand.Float64()*5 - 2.5,
			High:          price + rand.Float64()*3,
			Low:           prevClose - rand.Float64()*2,
			VWAP:          (prevClose + price) / 2,
			UpdatedAt:     time.Now().UnixNano(),
		}
	}

	// Sort by change percent descending
	for i := 0; i < len(gainers)-1; i++ {
		for j := i + 1; j < len(gainers); j++ {
			if gainers[j].ChangePercent > gainers[i].ChangePercent {
				gainers[i], gainers[j] = gainers[j], gainers[i]
			}
		}
	}

	return gainers
}

// SimulatePriceChange slightly modifies prices for real-time effect
func SimulatePriceChange(gainers []models.Gainer) []models.Gainer {
	rand.Seed(time.Now().UnixNano())

	updated := make([]models.Gainer, len(gainers))
	copy(updated, gainers)

	for i := range updated {
		// Random price fluctuation of -0.5% to +0.5%
		fluctuation := 1 + (rand.Float64()-0.5)*0.01
		updated[i].Price *= fluctuation
		updated[i].Change = updated[i].Price - updated[i].PreviousClose
		updated[i].ChangePercent = (updated[i].Change / updated[i].PreviousClose) * 100
		updated[i].UpdatedAt = time.Now().UnixNano()
	}

	return updated
}
