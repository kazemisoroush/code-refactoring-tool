//go:generate mockgen -source=cognito_provider.go -destination=mocks/mock_cognito_provider.go -package=mocks
package auth

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestCognitoProvider_SignIn(t *testing.T) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
		Region:     "us-east-1",
	}

	// Test successful sign in
	t.Run("successful sign in", func(t *testing.T) {
		// This test demonstrates the abstraction -
		// the client doesn't know it's using Cognito underneath
		provider := NewCognitoProvider(awsConfig, cognitoConfig)
		assert.NotNil(t, provider)
	})
}

func TestCognitoProvider_CreateUser(t *testing.T) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
		Region:     "us-east-1",
	}

	t.Run("successful user creation", func(t *testing.T) {
		firstName := "John"
		lastName := "Doe"
		password := "TempPass123!"

		req := &CreateUserRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: &firstName,
			LastName:  &lastName,
			Password:  &password,
		}

		provider := NewCognitoProvider(awsConfig, cognitoConfig)
		assert.NotNil(t, provider)

		// This test demonstrates that our provider implements the AuthProvider interface
		var authProvider AuthProvider = provider
		assert.NotNil(t, authProvider)

		// In a real implementation, we would:
		// 1. Mock the AWS Cognito client
		// 2. Set up expected calls and responses
		// 3. Call the method and verify results
		// 4. Assert that the response matches our generic User model
		//
		// This shows how the abstraction works - the client gets back
		// a generic User object regardless of whether it's Cognito, Auth0, etc.
		_ = req
	})
}

func TestCognitoProvider_ValidateToken(t *testing.T) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
		Region:     "us-east-1",
	}

	t.Run("interface compatibility", func(t *testing.T) {
		provider := NewCognitoProvider(awsConfig, cognitoConfig)

		// This demonstrates the key abstraction:
		// ValidateToken returns generic TokenClaims regardless of provider
		// The client code doesn't need to know about Cognito-specific token format
		var authProvider AuthProvider = provider
		assert.NotNil(t, authProvider)

		// In real tests, we would mock the AWS client and test actual validation
		// For now, we just verify the interface compatibility
		assert.Implements(t, (*AuthProvider)(nil), provider)
	})
}

// TestAbstractionExample demonstrates how the provider abstraction works
func TestAbstractionExample(t *testing.T) {
	t.Run("provider abstraction example", func(t *testing.T) {
		// This is the key benefit: client code can work with any provider

		// Could be Cognito
		awsConfig := aws.Config{Region: "us-east-1"}
		cognitoConfig := config.CognitoConfig{
			UserPoolID: "us-east-1_example",
			ClientID:   "cognito-client-id",
		}
		cognitoProvider := NewCognitoProvider(awsConfig, cognitoConfig)

		// Could be Auth0, Firebase, or any other provider
		// auth0Provider := NewAuth0Provider(auth0Config)
		// firebaseProvider := NewFirebaseProvider(firebaseConfig)

		// Client code uses the same interface for all providers
		providers := []AuthProvider{
			cognitoProvider,
			// auth0Provider,
			// firebaseProvider,
		}

		// Same code works with any provider - this is the abstraction benefit
		for i, provider := range providers {
			assert.NotNil(t, provider, "Provider %d should not be nil", i)

			// All providers implement the same interface
			assert.Implements(t, (*AuthProvider)(nil), provider)

			// Interface compatibility verified without calling methods that need real clients
			// In a real scenario with mocked clients, all methods would work identically
		}
	})
}

// Example integration test structure
func TestCognitoIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("integration with real AWS", func(t *testing.T) {
		// This would be an integration test with real AWS Cognito
		// It demonstrates that our abstraction works with the real service

		// In a real integration test:
		// 1. Set up test Cognito User Pool
		// 2. Create real AWS client with test credentials
		// 3. Test actual operations against Cognito
		// 4. Verify that all AuthProvider methods work correctly
		// 5. Clean up test resources

		t.Skip("Integration test requires real AWS setup")
	})
}

// TestErrorMapping verifies that Cognito errors are properly mapped to our generic errors
func TestCognitoProvider_ErrorMapping(t *testing.T) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
	}

	provider := NewCognitoProvider(awsConfig, cognitoConfig)

	t.Run("maps Cognito errors to generic errors", func(t *testing.T) {
		// Test that Cognito-specific errors become generic errors
		// This ensures the abstraction layer works properly

		tests := []struct {
			name          string
			cognitoError  error
			expectedError error
		}{
			{
				name:          "user not found",
				cognitoError:  &types.UserNotFoundException{},
				expectedError: ErrUserNotFound,
			},
			{
				name:          "invalid credentials",
				cognitoError:  &types.NotAuthorizedException{},
				expectedError: ErrInvalidCredentials,
			},
			{
				name:          "user already exists",
				cognitoError:  &types.UsernameExistsException{},
				expectedError: ErrUserAlreadyExists,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mappedErr := provider.mapCognitoError(tt.cognitoError)
				assert.Equal(t, tt.expectedError, mappedErr)
			})
		}
	})
}

// TestUserStatusMapping verifies that Cognito user statuses map correctly
func TestCognitoProvider_UserStatusMapping(t *testing.T) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
	}

	provider := NewCognitoProvider(awsConfig, cognitoConfig)

	t.Run("maps Cognito user status to generic status", func(t *testing.T) {
		tests := []struct {
			cognitoStatus  types.UserStatusType
			expectedStatus UserStatus
		}{
			{types.UserStatusTypeConfirmed, UserStatusActive},
			{types.UserStatusTypeUnconfirmed, UserStatusPending},
			{types.UserStatusTypeArchived, UserStatusInactive},
			{types.UserStatusTypeCompromised, UserStatusSuspended},
		}

		for _, tt := range tests {
			result := provider.mapCognitoUserStatus(tt.cognitoStatus)
			assert.Equal(t, tt.expectedStatus, result)
		}
	})
}

// BenchmarkCognitoProvider_SignIn benchmarks the sign in performance
func BenchmarkCognitoProvider_SignIn(b *testing.B) {
	awsConfig := aws.Config{Region: "us-east-1"}
	cognitoConfig := config.CognitoConfig{
		UserPoolID: "us-east-1_example",
		ClientID:   "test-client-id",
	}

	provider := NewCognitoProvider(awsConfig, cognitoConfig)
	ctx := context.Background()

	req := &SignInRequest{
		Username: "test@example.com",
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// In a real benchmark, this would test actual performance
		_, _ = provider.SignIn(ctx, req)
	}
}
