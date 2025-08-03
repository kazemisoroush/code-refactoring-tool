package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHealthController(t *testing.T) {
	serviceName := "test-service"
	version := "1.0.0"

	healthService := services.NewDefaultHealthService(serviceName, version)
	controller := NewHealthController(healthService)

	assert.NotNil(t, controller)
}

func TestHealthController_HealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	serviceName := "code-refactor-tool-api"
	version := "1.0.0"

	healthService := services.NewDefaultHealthService(serviceName, version)
	controller := NewHealthController(healthService)

	// Create a test router
	router := gin.New()
	router.GET("/health", controller.HealthCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, serviceName, response["service"])
	assert.Equal(t, version, response["version"])
	assert.Contains(t, response, "uptime")
}

func TestHealthController_HealthCheck_WithMetrics(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	serviceName := "code-refactor-tool-api"
	version := "1.0.0"

	healthService := services.NewDefaultHealthService(serviceName, version)
	controller := NewHealthController(healthService)

	// Create metrics middleware (disabled for testing)
	metricsMiddleware, err := middleware.NewMetricsMiddleware(config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     false, // Disabled for testing
	})
	require.NoError(t, err)

	// Create a test router with metrics middleware
	router := gin.New()
	router.Use(metricsMiddleware.Handle())
	router.GET("/health", controller.HealthCheck)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, serviceName, response["service"])
	assert.Equal(t, version, response["version"])
	assert.Contains(t, response, "uptime")
}
