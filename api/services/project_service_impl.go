// Package services provides business logic implementations
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
)

// ProjectServiceImpl provides project management functionality
type ProjectServiceImpl struct {
	projectRepo  repository.ProjectRepository
	codebaseRepo repository.CodebaseRepository
	taskRepo     repository.TaskRepository
}

// NewProjectService creates a new project service with dependency injection
func NewProjectService(
	projectRepo repository.ProjectRepository,
	codebaseRepo repository.CodebaseRepository,
	taskRepo repository.TaskRepository,
) ProjectService {
	return &ProjectServiceImpl{
		projectRepo:  projectRepo,
		codebaseRepo: codebaseRepo,
		taskRepo:     taskRepo,
	}
}

// CreateProject creates a new project
func (s *ProjectServiceImpl) CreateProject(ctx context.Context, request models.CreateProjectRequest) (*models.CreateProjectResponse, error) {
	slog.Info("Creating project", "name", request.Name)

	// Generate simple project ID
	projectID := s.generateProjectID()

	// Create project record
	project := &repository.ProjectRecord{
		ProjectID:   projectID,
		Name:        request.Name,
		Description: request.Description,
		Language:    request.Language,
		Tags:        request.Tags,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Persist project
	err := s.projectRepo.CreateProject(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &models.CreateProjectResponse{
		ProjectID: project.ProjectID,
		CreatedAt: project.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetProject retrieves a project
func (s *ProjectServiceImpl) GetProject(ctx context.Context, projectID string) (*models.GetProjectResponse, error) {
	slog.Info("Retrieving project", "project_id", projectID)

	// Retrieve project
	project, err := s.projectRepo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Build response
	response := &models.GetProjectResponse{
		ProjectID:   project.ProjectID,
		Name:        project.Name,
		Description: project.Description,
		Language:    project.Language,
		CreatedAt:   project.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   project.UpdatedAt.Format(time.RFC3339),
		Tags:        project.Tags,
		Metadata:    project.Metadata,
	}

	return response, nil
}

// UpdateProject updates a project
func (s *ProjectServiceImpl) UpdateProject(ctx context.Context, request models.UpdateProjectRequest) (*models.UpdateProjectResponse, error) {
	slog.Info("Updating project", "project_id", request.ProjectID)

	// Retrieve existing project
	existingProject, err := s.projectRepo.GetProject(ctx, request.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing project: %w", err)
	}

	// Build update
	updatedProject := s.buildProjectUpdate(existingProject, request)

	// Apply update
	err = s.projectRepo.UpdateProject(ctx, updatedProject)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &models.UpdateProjectResponse{
		ProjectID: updatedProject.ProjectID,
		UpdatedAt: updatedProject.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// DeleteProject deletes a project
func (s *ProjectServiceImpl) DeleteProject(ctx context.Context, projectID string) (*models.DeleteProjectResponse, error) {
	slog.Info("Deleting project", "project_id", projectID)

	// Delete the project
	if err := s.projectRepo.DeleteProject(ctx, projectID); err != nil {
		return nil, fmt.Errorf("failed to delete project: %w", err)
	}

	return &models.DeleteProjectResponse{
		Success: true,
	}, nil
}

// ListProjects lists projects
func (s *ProjectServiceImpl) ListProjects(ctx context.Context, request models.ListProjectsRequest) (*models.ListProjectsResponse, error) {
	slog.Info("Listing projects", "max_results", request.MaxResults)

	// Build repository options
	opts := repository.ListProjectsOptions{
		NextToken:  request.NextToken,
		MaxResults: request.MaxResults,
		TagFilter:  request.TagFilter,
	}

	// Retrieve projects
	projects, nextToken, err := s.projectRepo.ListProjects(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Build project summaries
	summaries := make([]models.ProjectSummary, len(projects))
	for i, project := range projects {
		summaries[i] = models.ProjectSummary{
			ProjectID: project.ProjectID,
			Name:      project.Name,
			CreatedAt: project.CreatedAt.Format(time.RFC3339),
			Tags:      project.Tags,
		}
	}

	response := &models.ListProjectsResponse{
		Projects: summaries,
	}

	if nextToken != "" {
		response.NextToken = &nextToken
	}

	return response, nil
}

// Helper Methods

// generateProjectID creates a simple project ID
func (s *ProjectServiceImpl) generateProjectID() string {
	return uuid.New().String()
}

// buildProjectUpdate creates an updated project model
func (s *ProjectServiceImpl) buildProjectUpdate(existing *repository.ProjectRecord, request models.UpdateProjectRequest) *repository.ProjectRecord {
	updated := *existing // Copy existing project
	updated.UpdatedAt = time.Now().UTC()

	// Apply selective updates
	if request.Name != nil {
		updated.Name = *request.Name
	}
	if request.Description != nil {
		updated.Description = request.Description
	}
	if request.Language != nil {
		updated.Language = request.Language
	}
	if request.Tags != nil {
		updated.Tags = request.Tags
	}
	if request.Metadata != nil {
		// Merge metadata instead of replacing
		if updated.Metadata == nil {
			updated.Metadata = make(map[string]string)
		}
		for k, v := range request.Metadata {
			updated.Metadata[k] = v
		}
	}

	return &updated
}
