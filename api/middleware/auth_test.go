package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthMiddleware(t *testing.T) {
	config := config.CognitoConfig{
		UserPoolID: "us-east-1_123456789",
		Region:     "us-east-1",
		ClientID:   "test-client-id",
	}

	middleware := NewAuthMiddleware(config)

	assert.NotNil(t, middleware)
	// Test that middleware was created successfully - interface is private, so basic functionality test
	assert.IsType(t, &AuthMiddleware{}, middleware)
}

func TestAuthMiddleware_RequireAuth_PublicEndpoints(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"health endpoint", "/health"},
		{"swagger endpoint", "/swagger"},
		{"docs endpoint", "/docs"},
		{"api-docs endpoint", "/api-docs"},
		{"swagger ui", "/swagger/index.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(config.CognitoConfig{})

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.Handle())
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestAuthMiddleware_RequireAuth_MissingAuthorizationHeader(t *testing.T) {
	middleware := NewAuthMiddleware(config.CognitoConfig{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response.Message)
}

func TestAuthMiddleware_RequireAuth_InvalidAuthorizationFormat(t *testing.T) {
	middleware := NewAuthMiddleware(config.CognitoConfig{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "invalid-token"},
		{"wrong prefix", "Basic invalid-token"},
		{"only Bearer", "Bearer"},
		{"empty value", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// NOTE: Additional whitebox tests for IsPublicEndpoint, TokenValidation, and JWK conversion
// are deferred until more comprehensive test infrastructure is established
