// Package repository provides user repository interfaces and implementations
package repository

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// UserRepository defines operations for user data storage
//
//go:generate mockgen -source=user_repository.go -destination=mocks/mock_user_repository.go -package=mocks
type UserRepository interface {
	// Basic CRUD operations
	CreateUser(ctx context.Context, user *models.DBUser) (*models.DBUser, error)
	GetUser(ctx context.Context, userID string) (*models.DBUser, error)
	GetUserByAuthID(ctx context.Context, authID string) (*models.DBUser, error)
	GetUserByEmail(ctx context.Context, email string) (*models.DBUser, error)
	UpdateUser(ctx context.Context, user *models.DBUser) (*models.DBUser, error)
	DeleteUser(ctx context.Context, userID string) error

	// Query operations
	ListUsers(ctx context.Context, filter *ListUsersFilter) ([]*models.DBUser, int, error)
	UserExists(ctx context.Context, userID string) (bool, error)

	// Project access operations
	HasProjectAccess(ctx context.Context, userID, projectID, accessType string) (bool, error)
	GrantProjectAccess(ctx context.Context, userID, projectID, accessType string) error
	RevokeProjectAccess(ctx context.Context, userID, projectID string) error
}

// ListUsersFilter defines filtering options for listing users
type ListUsersFilter struct {
	Limit  int                `validate:"omitempty,min=1,max=100"`
	Offset int                `validate:"omitempty,min=0"`
	Role   *models.UserRole   `validate:"omitempty"`
	Status *models.UserStatus `validate:"omitempty"`
	Search string             `validate:"omitempty,max=100"`
}
