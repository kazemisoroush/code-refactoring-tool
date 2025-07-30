// Package services provides concrete implementations for business logic
package services

import (
	"context"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// DefaultHealthService is the default implementation of HealthService
type DefaultHealthService struct {
	serviceName string
	version     string
	startTime   time.Time
}

// NewDefaultHealthService creates a new instance of DefaultHealthService
func NewDefaultHealthService(serviceName, version string) HealthService {
	return &DefaultHealthService{
		serviceName: serviceName,
		version:     version,
		startTime:   time.Now(),
	}
}

// GetHealthStatus retrieves the current health status of the service
func (s *DefaultHealthService) GetHealthStatus(_ context.Context, _ models.HealthCheckRequest) (*models.HealthCheckResponse, error) {
	// Calculate uptime
	uptime := int64(time.Since(s.startTime).Seconds())

	response := &models.HealthCheckResponse{
		Status:    string(models.HealthStatusHealthy),
		Service:   s.serviceName,
		Version:   s.version,
		Timestamp: time.Now().UTC(),
		Uptime:    &uptime,
	}

	return response, nil
}
