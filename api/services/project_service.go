// Package services provides business logic for the API layer
package services

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// ProjectService defines the interface for project-related operations
//
//go:generate mockgen -destination=./mocks/mock_project_service.go -mock_names=ProjectService=MockProjectService -package=mocks . ProjectService
type ProjectService interface {
	// CreateProject creates a new project with the given parameters
	CreateProject(ctx context.Context, request models.CreateProjectRequest) (*models.CreateProjectResponse, error)

	// GetProject retrieves a project by ID
	GetProject(ctx context.Context, projectID string) (*models.GetProjectResponse, error)

	// UpdateProject updates an existing project
	UpdateProject(ctx context.Context, request models.UpdateProjectRequest) (*models.UpdateProjectResponse, error)

	// DeleteProject deletes a project by ID
	DeleteProject(ctx context.Context, projectID string) (*models.DeleteProjectResponse, error)

	// ListProjects lists projects with pagination and filtering
	ListProjects(ctx context.Context, request models.ListProjectsRequest) (*models.ListProjectsResponse, error)
}
