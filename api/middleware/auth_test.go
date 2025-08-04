package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	assert.NotNil(t, middleware)
	authMiddleware, ok := middleware.(*AuthMiddleware)
	assert.True(t, ok)
	assert.Equal(t, mockProvider, authMiddleware.authProvider)
}

func TestAuthMiddleware_Handle_PublicEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)

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
			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			// Add middleware and a test handler
			router.Use(middleware.Handle())
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestAuthMiddleware_Handle_MissingAuthorizationHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	router.Use(middleware.Handle())
	router.GET("/api/projects", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Handle_InvalidAuthorizationFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		header string
	}{
		{"missing Bearer prefix", "token123"},
		{"wrong prefix", "Basic token123"},
		{"only Bearer", "Bearer"},
		{"empty value", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			router.Use(middleware.Handle())
			router.GET("/api/projects", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", "/api/projects", nil)
			req.Header.Set("Authorization", tt.header)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware_Handle_ValidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	// Mock successful token validation
	expectedClaims := &auth.TokenClaims{
		UserID:    "user123",
		Email:     "test@example.com",
		Username:  "testuser",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	mockProvider.EXPECT().ValidateToken(gomock.Any(), "valid-token").Return(expectedClaims, nil)

	router.Use(middleware.Handle())
	router.GET("/api/projects", func(c *gin.Context) {
		// Check that user information is stored in context
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "user123", userID)

		username, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)

		email, exists := c.Get("email")
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", email)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_Handle_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	// Mock token validation failure
	mockProvider.EXPECT().ValidateToken(gomock.Any(), "invalid-token").Return(nil, errors.New("token expired"))

	router.Use(middleware.Handle())
	router.GET("/api/projects", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_IsPublicEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	middleware := NewAuthMiddleware(mockProvider).(*AuthMiddleware)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/check", true},
		{"/swagger", true},
		{"/swagger/index.html", true},
		{"/docs", true},
		{"/docs/swagger.json", true},
		{"/api-docs", true},
		{"/api-docs/swagger.yaml", true},
		{"/api/projects", false},
		{"/api/agents", false},
		{"/v1/health", false}, // doesn't start with /health
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := middleware.isPublicEndpoint(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
