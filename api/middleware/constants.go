// Package middleware provides HTTP middleware components for the API
package middleware

// PublicEndpoints defines the list of API endpoints that don't require authentication
var PublicEndpoints = []string{
	"/health",
	"/swagger",
	"/docs",
	"/api-docs",
}
