// Package middleware provides an interface for middleware layers in a web application.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// Middleware is an interface for middleware layers
//
//go:generate mockgen -destination=./mocks/mock_middleware.go -mock_names=Middleware=MockMiddleware -package=mocks . Middleware
type Middleware interface {
	// Handle returns a gin.HandlerFunc that implements the middleware logic
	Handle() gin.HandlerFunc
}
