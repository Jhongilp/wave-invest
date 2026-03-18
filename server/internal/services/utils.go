package services

import "time"

// ParseTimeOrDefault parses a date string or returns a time N days ago
func ParseTimeOrDefault(dateStr string, defaultDaysAgo int) time.Time {
	if dateStr == "" {
		return time.Now().AddDate(0, 0, -defaultDaysAgo)
	}

	// Try parsing as RFC3339
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t
	}

	// Try parsing as date only
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t
	}

	// Default to N days ago
	return time.Now().AddDate(0, 0, -defaultDaysAgo)
}
