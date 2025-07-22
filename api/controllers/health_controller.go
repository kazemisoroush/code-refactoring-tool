// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
)

// HealthController handles health check requests
type HealthController struct {
	serviceName string
	version     string
}

// NewHealthController creates a new HealthController
func NewHealthController(serviceName, version string) *HealthController {
	return &HealthController{
		serviceName: serviceName,
		version:     version,
	}
}

// HealthCheck handles GET /health
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Service is healthy"
// @Router /health [get]
func (h *HealthController) HealthCheck(c *gin.Context) {
	// Send custom metric for health check
	if metrics := middleware.GetMetricsFromContext(c); metrics != nil {
		_ = metrics.SendCustomMetric("HealthCheck", 1, "Count", map[string]string{
			"Status": "healthy",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": h.serviceName,
		"version": h.version,
	})
}
