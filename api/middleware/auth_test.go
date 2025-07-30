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
	// TODO: Fix this assertion once the Middleware interface is implemented
	// assert.Equal(t, config, middleware.config)
	// assert.Equal(t, "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_123456789/.well-known/jwks.json", middleware.jwksURL)
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

// TODO: Fix these tests as they are whitebox tests and require to be better tested.
// func TestAuthMiddleware_IsPublicEndpoint(t *testing.T) {
// 	middleware := NewAuthMiddleware(config.CognitoConfig{})

// 	tests := []struct {
// 		path     string
// 		expected bool
// 	}{
// 		{"/health", true},
// 		{"/health/check", true},
// 		{"/swagger", true},
// 		{"/swagger/index.html", true},
// 		{"/docs", true},
// 		{"/docs/api", true},
// 		{"/api-docs", true},
// 		{"/api-docs/swagger.json", true},
// 		{"/agent/create", false},
// 		{"/agents", false},
// 		{"/protected", false},
// 		{"", false},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.path, func(t *testing.T) {
// 			result := middleware.isPublicEndpoint(tt.path)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }

// func TestCognitoClaims_TokenValidation(t *testing.T) {
// 	config := config.CognitoConfig{
// 		UserPoolID: "us-east-1_123456789",
// 		Region:     "us-east-1",
// 		ClientID:   "test-client-id",
// 	}

// 	middleware := NewAuthMiddleware(config)

// 	// Test with invalid token
// 	invalidToken := "invalid.token.here"
// 	_, err := middleware.validateToken(invalidToken)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "failed to parse token")
// }

// func TestJWKToRSAPublicKey(t *testing.T) {
// 	middleware := NewAuthMiddleware(config.CognitoConfig{})

// 	// Test with valid JWK (these are example values, not real keys)
// 	jwk := &JWK{
// 		Kty: "RSA",
// 		Kid: "test-kid",
// 		Use: "sig",
// 		N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbIS",
// 		E:   "AQAB",
// 	}

// 	publicKey, err := middleware.jwkToRSAPublicKey(jwk)
// 	require.NoError(t, err)
// 	assert.NotNil(t, publicKey)
// 	assert.NotNil(t, publicKey.N)
// 	assert.Equal(t, 65537, publicKey.E) // AQAB in base64 is 65537
// }

// func TestJWKToRSAPublicKey_InvalidBase64(t *testing.T) {
// 	middleware := NewAuthMiddleware(config.CognitoConfig{})

// 	// Test with invalid base64 in modulus
// 	jwk := &JWK{
// 		Kty: "RSA",
// 		N:   "invalid-base64-!!!",
// 		E:   "AQAB",
// 	}

// 	_, err := middleware.jwkToRSAPublicKey(jwk)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "failed to decode modulus")

// 	// Test with invalid base64 in exponent
// 	jwk = &JWK{
// 		Kty: "RSA",
// 		N:   "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbIS",
// 		E:   "invalid-base64-!!!",
// 	}

// 	_, err = middleware.jwkToRSAPublicKey(jwk)
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "failed to decode exponent")
// }

// func TestRespondWithError(t *testing.T) {
// 	middleware := NewAuthMiddleware(config.CognitoConfig{})

// 	gin.SetMode(gin.TestMode)
// 	w := httptest.NewRecorder()
// 	c, _ := gin.CreateTestContext(w)

// 	middleware.respondWithError(c, http.StatusUnauthorized, "Test error message")

// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// 	assert.True(t, c.IsAborted())

// 	var response models.ErrorResponse
// 	err := json.Unmarshal(w.Body.Bytes(), &response)
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusUnauthorized, response.Code)
// 	assert.Equal(t, "Test error message", response.Message)
// }
