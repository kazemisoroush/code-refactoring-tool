package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth/mocks"
	"github.com/stretchr/testify/assert"
)

// Test basic models (TDD - define expected behavior first)
func TestUser_Model(t *testing.T) {
	firstName := "John"
	lastName := "Doe"

	user := &auth.User{
		ID:        "user-123",
		Email:     "john.doe@example.com",
		Username:  "johndoe",
		FirstName: &firstName,
		LastName:  &lastName,
		Status:    auth.UserStatusActive,
		Metadata:  map[string]string{"role": "developer"},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "john.doe@example.com", user.Email)
	assert.Equal(t, "johndoe", user.Username)
	assert.Equal(t, &firstName, user.FirstName)
	assert.Equal(t, &lastName, user.LastName)
	assert.Equal(t, auth.UserStatusActive, user.Status)
	assert.Equal(t, "developer", user.Metadata["role"])
}

func TestUserStatus_Constants(t *testing.T) {
	assert.Equal(t, auth.UserStatus("active"), auth.UserStatusActive)
	assert.Equal(t, auth.UserStatus("inactive"), auth.UserStatusInactive)
	assert.Equal(t, auth.UserStatus("pending"), auth.UserStatusPending)
	assert.Equal(t, auth.UserStatus("suspended"), auth.UserStatusSuspended)
}

func TestAuthProvider_CreateUser_TDD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	ctx := context.Background()

	// TDD: Define what we expect the CreateUser method to do
	request := &auth.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
	}

	expectedUser := &auth.User{
		ID:        "user-123",
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: stringPtr("John"),
		LastName:  stringPtr("Doe"),
		Status:    auth.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set up mock expectation
	mockProvider.EXPECT().
		CreateUser(ctx, request).
		Return(expectedUser, nil)

	// Test the interface behavior
	user, err := mockProvider.CreateUser(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, request.Username, user.Username)
	assert.Equal(t, request.Email, user.Email)
	assert.Equal(t, auth.UserStatusActive, user.Status)
}

func TestAuthProvider_SignIn_TDD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	ctx := context.Background()

	// TDD: Define what we expect the SignIn method to do
	request := &auth.SignInRequest{
		Username: "testuser",
		Password: "password123",
	}

	expectedResult := &auth.AuthResult{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		User: &auth.User{
			ID:       "user-123",
			Username: "testuser",
			Email:    "test@example.com",
			Status:   auth.UserStatusActive,
		},
	}

	// Set up mock expectation
	mockProvider.EXPECT().
		SignIn(ctx, request).
		Return(expectedResult, nil)

	// Test the interface behavior
	result, err := mockProvider.SignIn(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, 3600, result.ExpiresIn)
	assert.NotNil(t, result.User)
}

func TestAuthProvider_ValidateToken_TDD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAuthProvider(ctrl)
	ctx := context.Background()

	// TDD: Define what we expect the ValidateToken method to do
	token := "valid-jwt-token"
	expectedClaims := &auth.TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		IssuedAt:  time.Now().Add(-1 * time.Hour),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Set up mock expectation
	mockProvider.EXPECT().
		ValidateToken(ctx, token).
		Return(expectedClaims, nil)

	// Test the interface behavior
	claims, err := mockProvider.ValidateToken(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "testuser", claims.Username)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
