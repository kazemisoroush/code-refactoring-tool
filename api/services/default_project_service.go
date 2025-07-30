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

// DefaultProjectService is the default implementation of ProjectService
type DefaultProjectService struct {
	projectRepo repository.ProjectRepository
}

// NewDefaultProjectService creates a new DefaultProjectService
func NewDefaultProjectService(projectRepo repository.ProjectRepository) *DefaultProjectService {
	return &DefaultProjectService{
		projectRepo: projectRepo,
	}
}

// CreateProject creates a new project with the given parameters
func (s *DefaultProjectService) CreateProject(ctx context.Context, request models.CreateProjectRequest) (*models.CreateProjectResponse, error) {
	// Generate a unique project ID
	projectID := generateProjectID()

	// Create project record
	now := time.Now().UTC()
	projectRecord := &repository.ProjectRecord{
		ProjectID:   projectID,
		Name:        request.Name,
		Description: request.Description,
		Language:    request.Language,
		Status:      string(models.ProjectStatusActive),
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        request.Tags,
		Metadata:    make(map[string]string),
	}

	// Store in repository
	if err := s.projectRepo.CreateProject(ctx, projectRecord); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &models.CreateProjectResponse{
		ProjectID: projectID,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}

// GetProject retrieves a project by ID
func (s *DefaultProjectService) GetProject(ctx context.Context, projectID string) (*models.GetProjectResponse, error) {
	projectRecord, err := s.projectRepo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if projectRecord == nil {
		return nil, fmt.Errorf("project not found")
	}

	return projectRecord.ToGetProjectResponse(), nil
}

// UpdateProject updates an existing project
func (s *DefaultProjectService) UpdateProject(ctx context.Context, request models.UpdateProjectRequest) (*models.UpdateProjectResponse, error) {
	// Get existing project
	projectRecord, err := s.projectRepo.GetProject(ctx, request.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if projectRecord == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Update fields if provided
	if request.Name != nil {
		projectRecord.Name = *request.Name
	}
	if request.Description != nil {
		projectRecord.Description = request.Description
	}
	if request.Language != nil {
		projectRecord.Language = request.Language
	}
	if request.Tags != nil {
		projectRecord.Tags = request.Tags
	}
	if request.Metadata != nil {
		projectRecord.Metadata = request.Metadata
	}

	// Update timestamp
	projectRecord.UpdatedAt = time.Now().UTC()

	// Store in repository
	if err := s.projectRepo.UpdateProject(ctx, projectRecord); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &models.UpdateProjectResponse{
		ProjectID: request.ProjectID,
		UpdatedAt: projectRecord.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteProject deletes a project by ID
func (s *DefaultProjectService) DeleteProject(ctx context.Context, projectID string) (*models.DeleteProjectResponse, error) {
	// Check if project exists
	exists, err := s.projectRepo.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to check project existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("project not found")
	}

	// Delete project
	if err := s.projectRepo.DeleteProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return &models.DeleteProjectResponse{
		Success: true,
	}, nil
}

// ListProjects lists projects with pagination and filtering
func (s *DefaultProjectService) ListProjects(ctx context.Context, request models.ListProjectsRequest) (*models.ListProjectsResponse, error) {
	opts := repository.ListProjectsOptions{
		NextToken:  request.NextToken,
		MaxResults: request.MaxResults,
		TagFilter:  request.TagFilter,
	}

	projectRecords, nextToken, err := s.projectRepo.ListProjects(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Convert to response format
	projects := make([]models.ProjectSummary, len(projectRecords))
	for i, record := range projectRecords {
		projects[i] = record.ToProjectSummary()
	}

	response := &models.ListProjectsResponse{
		Projects: projects,
	}

	if nextToken != "" {
		response.NextToken = &nextToken
	}

	return response, nil
}

// generateProjectID generates a unique project ID
func generateProjectID() string {
	// Generate UUID and format as AWS-style resource identifier
	id := uuid.New().String()
	return fmt.Sprintf("proj-%s", id[:13])
}
