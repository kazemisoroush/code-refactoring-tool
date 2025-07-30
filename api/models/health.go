// Package models provides data transfer objects and request/response models for the API
package models

import (
	"time"
)

// HealthCheckRequest represents the request for health check (no parameters needed)
type HealthCheckRequest struct {
	// No parameters needed for health check
} //@name HealthCheckRequest

// HealthCheckResponse represents the response for health check
type HealthCheckResponse struct {
	// Service status
	Status string `json:"status" example:"healthy"`
	// Service name
	Service string `json:"service" example:"code-refactor-tool-api"`
	// Service version
	Version string `json:"version" example:"1.0.0"`
	// Current timestamp
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	// Uptime in seconds (optional)
	Uptime *int64 `json:"uptime,omitempty" example:"3600"`
} //@name HealthCheckResponse

// HealthStatus represents the possible health statuses
type HealthStatus string

// Health status constants
const (
	// HealthStatusHealthy indicates the service is healthy
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy indicates the service is unhealthy
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusDegraded indicates the service is partially functioning
	HealthStatusDegraded HealthStatus = "degraded"
)

// String returns the string representation of the health status
func (h HealthStatus) String() string {
	return string(h)
}
