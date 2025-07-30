// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// HealthController handles health check requests
type HealthController struct {
	healthService services.HealthService
}

// NewHealthController creates a new HealthController
func NewHealthController(healthService services.HealthService) *HealthController {
	return &HealthController{
		healthService: healthService,
	}
}

// HealthCheck handles GET /health
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthCheckResponse "Service is healthy"
// @Failure 500 {object} models.ErrorResponse "Service is unhealthy"
// @Router /health [get]
func (h *HealthController) HealthCheck(ctx *gin.Context) {
	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.HealthCheckRequest](ctx)
	if !exists {
		// Fall back to empty request for backward compatibility
		request = models.HealthCheckRequest{}
	}

	// Send custom metric for health check
	if metrics := middleware.GetMetricsFromContext(ctx); metrics != nil {
		_ = metrics.SendCustomMetric("HealthCheck", 1, "Count", map[string]string{
			"Status": string(models.HealthStatusHealthy),
		})
	}

	// Call the service to get health status
	response, err := h.healthService.GetHealthStatus(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get health status",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
