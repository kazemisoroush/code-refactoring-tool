// Package auth provides provider-agnostic authentication interfaces and models
package auth

import (
	"context"
	"time"
)

// AuthProvider defines the interface for authentication providers (Cognito, Auth0, Firebase, etc.)
//
//go:generate mockgen -source=provider.go -destination=mocks/mock_auth_provider.go -package=mocks
//nolint:revive // AuthProvider is a well-established name in the codebase
type AuthProvider interface {
	// User Management
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)

	// Authentication
	SignUp(ctx context.Context, req *SignUpRequest) (*AuthResult, error)
	SignIn(ctx context.Context, req *SignInRequest) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	SignOut(ctx context.Context, accessToken string) error

	// Email Confirmation
	ConfirmSignUp(ctx context.Context, username, confirmationCode string) error

	// Token Validation
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)

	// Password Management
	ResetPassword(ctx context.Context, email string) error
	ConfirmPasswordReset(ctx context.Context, req *PasswordResetRequest) error
}

// User represents a generic user from any auth provider
type User struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	Username  string            `json:"username"`
	FirstName *string           `json:"first_name,omitempty"`
	LastName  *string           `json:"last_name,omitempty"`
	Status    UserStatus        `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// UserStatus represents the status of a user
type UserStatus string

const (
	// UserStatusActive indicates an active user account
	UserStatusActive UserStatus = "active"
	// UserStatusInactive indicates an inactive user account
	UserStatusInactive UserStatus = "inactive"
	// UserStatusPending indicates a user account pending activation
	UserStatusPending UserStatus = "pending"
	// UserStatusSuspended indicates a suspended user account
	UserStatusSuspended UserStatus = "suspended"
)

// AuthResult contains authentication response data
//
//nolint:revive // AuthResult is a well-established name in the codebase
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user"`
}

// TokenClaims represents validated token claims
type TokenClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Password  *string `json:"password,omitempty"` // Optional for some providers
}

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Filter string `json:"filter,omitempty"`
}

// ListUsersResponse represents the response to a list users request
type ListUsersResponse struct {
	Users      []*User `json:"users"`
	TotalCount int     `json:"total_count"`
	NextToken  *string `json:"next_token,omitempty"`
}

// SignUpRequest represents a user registration request
type SignUpRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// SignInRequest represents a user authentication request
type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PasswordResetRequest represents the request to reset a user's password
type PasswordResetRequest struct {
	Email            string `json:"email"`
	ConfirmationCode string `json:"confirmation_code"`
	NewPassword      string `json:"new_password"`
}
