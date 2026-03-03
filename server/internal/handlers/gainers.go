package handlers

import (
	"net/http"

	"wave_invest/server/internal/models"

	"github.com/gin-gonic/gin"
)

// GainersGetter is an interface for retrieving gainers data
type GainersGetter interface {
	GetLastGainers() []models.Gainer
}

// GainersHandler handles HTTP requests for gainers data
type GainersHandler struct {
	getter GainersGetter
}

// NewGainersHandler creates a new gainers handler
func NewGainersHandler(getter GainersGetter) *GainersHandler {
	return &GainersHandler{getter: getter}
}

// GetGainers returns the current top gainers
func (h *GainersHandler) GetGainers(c *gin.Context) {
	gainers := h.getter.GetLastGainers()

	if gainers == nil {
		c.JSON(http.StatusOK, gin.H{
			"data":    []models.Gainer{},
			"message": "No data available yet. Waiting for first poll.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gainers,
	})
}

// HealthHandler returns server health status
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
