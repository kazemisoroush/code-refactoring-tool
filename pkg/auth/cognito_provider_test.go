// Package auth provides authentication providers for various services like AWS Cognito.
package auth

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
)

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
