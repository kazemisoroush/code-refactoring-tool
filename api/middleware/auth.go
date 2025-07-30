// Package middleware provides HTTP middleware components for the API
package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a set of JSON Web Keys
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// CognitoClaims represents the claims in a Cognito JWT token
type CognitoClaims struct {
	TokenUse string `json:"token_use"`
	Sub      string `json:"sub"`
	Username string `json:"cognito:username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides authentication middleware for Cognito JWT tokens
type AuthMiddleware struct {
	config  config.CognitoConfig
	jwkSet  *JWKSet
	jwksURL string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config config.CognitoConfig) Middleware {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		config.Region, config.UserPoolID)

	return &AuthMiddleware{
		config:  config,
		jwksURL: jwksURL,
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
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.respondWithError(c, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := tokenParts[1]

		// Validate the token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.respondWithError(c, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// Store user information in the context for downstream handlers
		c.Set("user_id", claims.Sub)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// validateToken validates a JWT token against Cognito's public keys
func (m *AuthMiddleware) validateToken(tokenString string) (*CognitoClaims, error) {
	// Parse the token without verification first to get the kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &CognitoClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get the key ID from the token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	// Get the public key for this kid
	publicKey, err := m.getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate the token with the public key
	claims := &CognitoClaims{}
	validToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token with claims: %w", err)
	}

	if !validToken.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	// Validate token type (should be "access" for API access)
	if claims.TokenUse != "access" {
		return nil, fmt.Errorf("invalid token use: %s", claims.TokenUse)
	}

	// Validate audience (client ID) - for Cognito access tokens, check 'aud' claim
	if claims.Audience == nil {
		return nil, fmt.Errorf("missing audience in token")
	}

	found := false
	for _, aud := range claims.Audience {
		if aud == m.config.ClientID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid audience")
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s",
		m.config.Region, m.config.UserPoolID)
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	// Validate expiration
	if claims.ExpiresAt != nil &&
		claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// getPublicKey retrieves the public key for the given key ID
func (m *AuthMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Lazy load JWK set if not already loaded
	if m.jwkSet == nil {
		if err := m.loadJWKSet(); err != nil {
			return nil, fmt.Errorf("failed to load JWK set: %w", err)
		}
	}

	// Find the key with the matching kid
	var jwk *JWK
	for _, key := range m.jwkSet.Keys {
		if key.Kid == kid {
			jwk = &key
			break
		}
	}

	if jwk == nil {
		return nil, fmt.Errorf("key not found for kid: %s", kid)
	}

	// Convert JWK to RSA public key
	return m.jwkToRSAPublicKey(jwk)
}

// loadJWKSet loads the JWK set from Cognito
func (m *AuthMiddleware) loadJWKSet() error {
	resp, err := http.Get(m.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWK set: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() // Ignore error in defer cleanup
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWK set: status %d", resp.StatusCode)
	}

	var jwkSet JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwkSet); err != nil {
		return fmt.Errorf("failed to decode JWK set: %w", err)
	}

	m.jwkSet = &jwkSet
	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func (m *AuthMiddleware) jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	// Create the RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	return publicKey, nil
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
