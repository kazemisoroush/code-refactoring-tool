// Package services provides business logic for the API layer
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
)

// DefaultCodebaseService is the default implementation of CodebaseService
type DefaultCodebaseService struct {
	codebaseRepo repository.CodebaseRepository
}

// NewDefaultCodebaseService creates a new DefaultCodebaseService
func NewDefaultCodebaseService(codebaseRepo repository.CodebaseRepository) *DefaultCodebaseService {
	return &DefaultCodebaseService{
		codebaseRepo: codebaseRepo,
	}
}

// CreateCodebase creates a new codebase with the given parameters
func (s *DefaultCodebaseService) CreateCodebase(ctx context.Context, request models.CreateCodebaseRequest) (*models.CreateCodebaseResponse, error) {
	// Generate a unique codebase ID
	codebaseID := uuid.New().String()

	// Create the codebase entity
	now := time.Now().UTC()
	codebase := &models.Codebase{
		CodebaseID:    codebaseID,
		ProjectID:     request.ProjectID,
		Name:          request.Name,
		Provider:      request.Provider,
		URL:           request.URL,
		DefaultBranch: request.DefaultBranch,
		CreatedAt:     now,
		UpdatedAt:     now,
		Tags:          request.Tags,
	}

	// Initialize empty metadata if nil
	if codebase.Tags == nil {
		codebase.Tags = make(map[string]string)
	}
	if codebase.Metadata == nil {
		codebase.Metadata = make(map[string]string)
	}

	// Save to repository
	if err := s.codebaseRepo.CreateCodebase(ctx, codebase); err != nil {
		return nil, fmt.Errorf("failed to create codebase: %w", err)
	}

	return &models.CreateCodebaseResponse{
		CodebaseID: codebaseID,
		CreatedAt:  now.Format(time.RFC3339),
	}, nil
}

// GetCodebase retrieves a codebase by ID
func (s *DefaultCodebaseService) GetCodebase(ctx context.Context, codebaseID string) (*models.GetCodebaseResponse, error) {
	codebase, err := s.codebaseRepo.GetCodebase(ctx, codebaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase: %w", err)
	}

	return &models.GetCodebaseResponse{
		CodebaseID:    codebase.CodebaseID,
		ProjectID:     codebase.ProjectID,
		Name:          codebase.Name,
		Provider:      codebase.Provider,
		URL:           codebase.URL,
		DefaultBranch: codebase.DefaultBranch,
		CreatedAt:     codebase.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     codebase.UpdatedAt.Format(time.RFC3339),
		Metadata:      codebase.Metadata,
		Tags:          codebase.Tags,
	}, nil
}

// UpdateCodebase updates an existing codebase
func (s *DefaultCodebaseService) UpdateCodebase(ctx context.Context, request models.UpdateCodebaseRequest) (*models.UpdateCodebaseResponse, error) {
	// Get the existing codebase
	codebase, err := s.codebaseRepo.GetCodebase(ctx, request.CodebaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase for update: %w", err)
	}

	// Update only the provided fields
	if request.Name != nil {
		codebase.Name = *request.Name
	}
	if request.DefaultBranch != nil {
		codebase.DefaultBranch = *request.DefaultBranch
	}
	if request.Tags != nil {
		codebase.Tags = request.Tags
	}
	if request.Metadata != nil {
		codebase.Metadata = request.Metadata
	}

	// Update the timestamp
	codebase.UpdatedAt = time.Now().UTC()

	// Save the updated codebase
	if err := s.codebaseRepo.UpdateCodebase(ctx, codebase); err != nil {
		return nil, fmt.Errorf("failed to update codebase: %w", err)
	}

	return &models.UpdateCodebaseResponse{
		CodebaseID: codebase.CodebaseID,
		UpdatedAt:  codebase.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteCodebase deletes a codebase by ID
func (s *DefaultCodebaseService) DeleteCodebase(ctx context.Context, codebaseID string) (*models.DeleteCodebaseResponse, error) {
	if err := s.codebaseRepo.DeleteCodebase(ctx, codebaseID); err != nil {
		return nil, fmt.Errorf("failed to delete codebase: %w", err)
	}

	return &models.DeleteCodebaseResponse{
		Success: true,
	}, nil
}

// ListCodebases lists codebases with pagination and filtering
func (s *DefaultCodebaseService) ListCodebases(ctx context.Context, request models.ListCodebasesRequest) (*models.ListCodebasesResponse, error) {
	// Build filter from request
	filter := repository.CodebaseFilter{
		ProjectID:  request.ProjectID,
		TagFilter:  request.TagFilter,
		NextToken:  request.NextToken,
		MaxResults: request.MaxResults,
	}

	// Get codebases from repository
	codebases, nextToken, err := s.codebaseRepo.ListCodebases(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list codebases: %w", err)
	}

	// Convert to response format
	summaries := make([]models.CodebaseSummary, 0, len(codebases))
	for _, codebase := range codebases {
		summary := models.CodebaseSummary{
			CodebaseID:    codebase.CodebaseID,
			ProjectID:     codebase.ProjectID,
			Name:          codebase.Name,
			Provider:      codebase.Provider,
			URL:           codebase.URL,
			DefaultBranch: codebase.DefaultBranch,
			CreatedAt:     codebase.CreatedAt.Format(time.RFC3339),
			Tags:          codebase.Tags,
		}
		summaries = append(summaries, summary)
	}

	response := &models.ListCodebasesResponse{
		Codebases: summaries,
	}

	if nextToken != "" {
		response.NextToken = &nextToken
	}

	return response, nil
}
