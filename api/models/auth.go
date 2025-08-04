// Package models provides authentication-related API models
package models

import "time"

// UserRole defines the role of a user in the system
type UserRole string

const (
	// RoleOwner has full access to everything
	RoleOwner UserRole = "owner" // Full access to everything
	// RoleAdmin can manage users and all projects
	RoleAdmin UserRole = "admin" // Manage users, all projects
	// RoleDeveloper can create and manage own projects
	RoleDeveloper UserRole = "developer" // Create/manage own projects
	// RoleViewer has read-only access
	RoleViewer UserRole = "viewer" // Read-only access
)

// UserStatus represents the current status of a user account
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

// Permission represents a specific permission in the system
type Permission string

const (
	// PermissionProjectCreate allows creating new projects
	PermissionProjectCreate Permission = "project:create"
	// PermissionProjectRead allows reading project information
	PermissionProjectRead Permission = "project:read"
	// PermissionProjectUpdate allows updating project information
	PermissionProjectUpdate Permission = "project:update"
	// PermissionProjectDelete allows deleting projects
	PermissionProjectDelete Permission = "project:delete"

	// PermissionTaskCreate allows creating new tasks
	PermissionTaskCreate Permission = "task:create"
	// PermissionTaskExecute allows executing tasks
	PermissionTaskExecute Permission = "task:execute"
	// PermissionTaskRead allows reading task information
	PermissionTaskRead Permission = "task:read"
	// PermissionTaskUpdate allows updating task information
	PermissionTaskUpdate Permission = "task:update"
	// PermissionTaskDelete allows deleting tasks
	PermissionTaskDelete Permission = "task:delete"

	// PermissionAgentCreate allows creating new agents
	PermissionAgentCreate Permission = "agent:create"
	// PermissionAgentRead allows reading agent information
	PermissionAgentRead Permission = "agent:read"
	// PermissionAgentUpdate allows updating agent information
	PermissionAgentUpdate Permission = "agent:update"
	// PermissionAgentDelete allows deleting agents
	PermissionAgentDelete Permission = "agent:delete"

	// PermissionUserCreate allows creating new users
	PermissionUserCreate Permission = "user:create"
	// PermissionUserRead allows reading user information
	PermissionUserRead Permission = "user:read"
	// PermissionUserUpdate allows updating user information
	PermissionUserUpdate Permission = "user:update"
	// PermissionUserDelete allows deleting users
	PermissionUserDelete Permission = "user:delete"
)

// API Request/Response Models

// SignUpRequest represents a user registration request
type SignUpRequest struct {
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,max=50"`
}

// SignUpResponse represents the response to a user registration request
type SignUpResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         *APIUser `json:"user"`
}

// SignInRequest represents a user authentication request
type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SignInResponse represents the response to a user authentication request
type SignInResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         *APIUser `json:"user"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the response to a token refresh request
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// SignOutRequest represents a user sign out request
type SignOutRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username  string   `json:"username" validate:"required,min=3,max=50"`
	Email     string   `json:"email" validate:"required,email"`
	FirstName *string  `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName  *string  `json:"last_name,omitempty" validate:"omitempty,max=50"`
	Role      UserRole `json:"role,omitempty" validate:"omitempty,oneof=owner admin developer viewer"`
}

// CreateUserResponse represents the response to a user creation request
type CreateUserResponse struct {
	User *APIUser `json:"user"`
}

// GetUserResponse represents the response to a get user request
type GetUserResponse struct {
	User *APIUser `json:"user"`
}

// UpdateUserRequest represents a request to update user information
type UpdateUserRequest struct {
	UserID    string    `json:"user_id" validate:"required"`
	Email     *string   `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string   `json:"first_name,omitempty" validate:"omitempty,max=50"`
	LastName  *string   `json:"last_name,omitempty" validate:"omitempty,max=50"`
	Role      *UserRole `json:"role,omitempty" validate:"omitempty,oneof=owner admin developer viewer"`
}

// UpdateUserResponse represents the response to a user update request
type UpdateUserResponse struct {
	User *APIUser `json:"user"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Limit  int         `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int         `json:"offset,omitempty" validate:"omitempty,min=0"`
	Role   *UserRole   `json:"role,omitempty" validate:"omitempty,oneof=owner admin developer viewer"`
	Status *UserStatus `json:"status,omitempty" validate:"omitempty,oneof=active inactive pending suspended"`
}

// ListUsersResponse represents the response to a list users request
type ListUsersResponse struct {
	Users  []*APIUser `json:"users"`
	Total  int        `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// APIUser represents a user model returned to API clients
type APIUser struct {
	UserID    string     `json:"user_id"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// UserContext represents user context for authenticated requests
type UserContext struct {
	UserID      string       `json:"user_id"`
	AuthID      string       `json:"auth_id"` // Provider-specific ID
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	Role        UserRole     `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// DBUser represents a user model for internal database storage
type DBUser struct {
	UserID    string     `json:"user_id" db:"user_id"`
	AuthID    string     `json:"auth_id" db:"auth_id"` // Provider-specific ID
	Email     string     `json:"email" db:"email"`
	Username  string     `json:"username" db:"username"`
	FirstName *string    `json:"first_name,omitempty" db:"first_name"`
	LastName  *string    `json:"last_name,omitempty" db:"last_name"`
	Role      UserRole   `json:"role" db:"role"`
	Status    UserStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}
