// Package services provides business logic for the API layer
package services

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// CodebaseService defines the interface for codebase-related operations
//
//go:generate mockgen -destination=./mocks/mock_codebase_service.go -mock_names=CodebaseService=MockCodebaseService -package=mocks . CodebaseService
type CodebaseService interface {
	// CreateCodebase creates a new codebase with the given parameters
	CreateCodebase(ctx context.Context, request models.CreateCodebaseRequest) (*models.CreateCodebaseResponse, error)

	// GetCodebase retrieves a codebase by ID
	GetCodebase(ctx context.Context, codebaseID string) (*models.GetCodebaseResponse, error)

	// UpdateCodebase updates an existing codebase
	UpdateCodebase(ctx context.Context, request models.UpdateCodebaseRequest) (*models.UpdateCodebaseResponse, error)

	// DeleteCodebase deletes a codebase by ID
	DeleteCodebase(ctx context.Context, codebaseID string) (*models.DeleteCodebaseResponse, error)

	// ListCodebases lists codebases with pagination and filtering
	ListCodebases(ctx context.Context, request models.ListCodebasesRequest) (*models.ListCodebasesResponse, error)
}
