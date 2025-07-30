// Package repository provides data access interfaces and implementations
package repository

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// CodebaseRepository defines the interface for codebase data persistence
//
//go:generate mockgen -destination=./mocks/mock_codebase_repository.go -mock_names=CodebaseRepository=MockCodebaseRepository -package=mocks . CodebaseRepository
type CodebaseRepository interface {
	// CreateCodebase creates a new codebase record
	CreateCodebase(ctx context.Context, codebase *models.Codebase) error

	// GetCodebase retrieves a codebase by ID
	GetCodebase(ctx context.Context, codebaseID string) (*models.Codebase, error)

	// UpdateCodebase updates an existing codebase
	UpdateCodebase(ctx context.Context, codebase *models.Codebase) error

	// DeleteCodebase deletes a codebase by ID
	DeleteCodebase(ctx context.Context, codebaseID string) error

	// ListCodebases lists codebases with optional filtering and pagination
	// Returns codebases and next token for pagination
	ListCodebases(ctx context.Context, filter CodebaseFilter) ([]*models.Codebase, string, error)

	// CodebaseExists checks if a codebase exists
	CodebaseExists(ctx context.Context, codebaseID string) (bool, error)

	// GetCodebasesByProject gets all codebases for a specific project
	GetCodebasesByProject(ctx context.Context, projectID string) ([]*models.Codebase, error)
}

// CodebaseFilter defines filtering options for listing codebases
type CodebaseFilter struct {
	ProjectID  *string
	Provider   *models.Provider
	TagFilter  *string
	NextToken  *string
	MaxResults *int
}
