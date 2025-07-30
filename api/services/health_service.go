// Package services provides business logic for the API layer
package services

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// HealthService defines the interface for health-related operations
//
//go:generate mockgen -destination=./mocks/mock_health_service.go -mock_names=HealthService=MockHealthService -package=mocks . HealthService
type HealthService interface {
	// GetHealthStatus retrieves the current health status of the service
	GetHealthStatus(ctx context.Context, request models.HealthCheckRequest) (*models.HealthCheckResponse, error)
}
