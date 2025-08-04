// Package middleware provides HTTP middleware components for the API
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
)

// AuthMiddleware provides provider-agnostic authentication middleware
// This middleware delegates token validation to the configured AuthProvider,
// allowing easy switching between Cognito, Auth0, Firebase, or any other provider
// that implements the AuthProvider interface.
type AuthMiddleware struct {
	authProvider auth.AuthProvider
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authProvider auth.AuthProvider) Middleware {
	return &AuthMiddleware{
		authProvider: authProvider,
	}
}

// Handle is the middleware function that validates JWT tokens
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for health check and swagger endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.respondWithError(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Check for Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" || strings.TrimSpace(tokenParts[1]) == "" {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := strings.TrimSpace(tokenParts[1])

		// Validate the token using the auth provider
		claims, err := m.authProvider.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid token: "+err.Error())
			return
		}

		// Store user information in the context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// isPublicEndpoint checks if the given path is a public endpoint
func (m *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/swagger",
		"/docs",
		"/api-docs",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

// respondWithError sends an error response
func (m *AuthMiddleware) respondWithError(c *gin.Context, statusCode int, message string) {
	errorResponse := models.ErrorResponse{
		Code:    statusCode,
		Message: message,
	}
	c.JSON(statusCode, errorResponse)
	c.Abort()
}
