package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthProvider is a mock implementation of auth.AuthProvider for testing
type MockAuthProvider struct {
	mock.Mock
}

func (m *MockAuthProvider) CreateUser(ctx context.Context, req *auth.CreateUserRequest) (*auth.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) GetUser(ctx context.Context, userID string) (*auth.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) UpdateUser(ctx context.Context, userID string, req *auth.UpdateUserRequest) (*auth.User, error) {
	args := m.Called(ctx, userID, req)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthProvider) ListUsers(ctx context.Context, req *auth.ListUsersRequest) (*auth.ListUsersResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*auth.ListUsersResponse), args.Error(1)
}

func (m *MockAuthProvider) SignUp(ctx context.Context, req *auth.SignUpRequest) (*auth.AuthResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*auth.AuthResult), args.Error(1)
}

func (m *MockAuthProvider) SignIn(ctx context.Context, req *auth.SignInRequest) (*auth.AuthResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*auth.AuthResult), args.Error(1)
}

func (m *MockAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthResult, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(*auth.AuthResult), args.Error(1)
}

func (m *MockAuthProvider) SignOut(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthProvider) ValidateToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *MockAuthProvider) ResetPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockAuthProvider) ConfirmPasswordReset(ctx context.Context, req *auth.PasswordResetRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestNewProviderAuthMiddleware(t *testing.T) {
	mockProvider := &MockAuthProvider{}
	middleware := NewAuthMiddleware(mockProvider)

	assert.NotNil(t, middleware)
	authMiddleware, ok := middleware.(*AuthMiddleware)
	assert.True(t, ok)
	assert.Equal(t, mockProvider, authMiddleware.authProvider)
}

func TestProviderAuthMiddleware_Handle_PublicEndpoints(t *testing.T) {
	mockProvider := &MockAuthProvider{}
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

func TestProviderAuthMiddleware_Handle_MissingAuthorizationHeader(t *testing.T) {
	mockProvider := &MockAuthProvider{}
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

func TestProviderAuthMiddleware_Handle_InvalidAuthorizationFormat(t *testing.T) {
	mockProvider := &MockAuthProvider{}
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

func TestProviderAuthMiddleware_Handle_ValidToken(t *testing.T) {
	mockProvider := &MockAuthProvider{}
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
	mockProvider.On("ValidateToken", mock.Anything, "valid-token").Return(expectedClaims, nil)

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
	mockProvider.AssertExpectations(t)
}

func TestProviderAuthMiddleware_Handle_InvalidToken(t *testing.T) {
	mockProvider := &MockAuthProvider{}
	middleware := NewAuthMiddleware(mockProvider)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, router := gin.CreateTestContext(w)

	// Mock token validation failure
	mockProvider.On("ValidateToken", mock.Anything, "invalid-token").Return(nil, errors.New("token expired"))

	router.Use(middleware.Handle())
	router.GET("/api/projects", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockProvider.AssertExpectations(t)
}

func TestAuthMiddleware_IsPublicEndpoint(t *testing.T) {
	mockProvider := &MockAuthProvider{}
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
