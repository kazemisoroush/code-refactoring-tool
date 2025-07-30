// Package repository provides data access layer for the API
package repository

import (
	"context"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// ProjectRecord represents the project data stored in the database
type ProjectRecord struct {
	ProjectID   string            `json:"project_id" db:"project_id"`
	Name        string            `json:"name" db:"name"`
	Description *string           `json:"description,omitempty" db:"description"`
	Language    *string           `json:"language,omitempty" db:"language"`
	Status      string            `json:"status" db:"status"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	Tags        map[string]string `json:"tags,omitempty" db:"tags"`
	Metadata    map[string]string `json:"metadata,omitempty" db:"metadata"`
}

// ToGetProjectResponse converts ProjectRecord to GetProjectResponse
func (r *ProjectRecord) ToGetProjectResponse() *models.GetProjectResponse {
	return &models.GetProjectResponse{
		ProjectID:   r.ProjectID,
		Name:        r.Name,
		Description: r.Description,
		Language:    r.Language,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
		Tags:        r.Tags,
		Metadata:    r.Metadata,
	}
}

// ToProjectSummary converts ProjectRecord to ProjectSummary
func (r *ProjectRecord) ToProjectSummary() models.ProjectSummary {
	return models.ProjectSummary{
		ProjectID: r.ProjectID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		Tags:      r.Tags,
	}
}

// ProjectRepository defines the interface for project data operations
//
//go:generate mockgen -destination=./mocks/mock_project_repository.go -mock_names=ProjectRepository=MockProjectRepository -package=mocks . ProjectRepository
type ProjectRepository interface {
	// CreateProject creates a new project record
	CreateProject(ctx context.Context, project *ProjectRecord) error

	// GetProject retrieves a project by ID
	GetProject(ctx context.Context, projectID string) (*ProjectRecord, error)

	// UpdateProject updates an existing project record
	UpdateProject(ctx context.Context, project *ProjectRecord) error

	// DeleteProject deletes a project by ID
	DeleteProject(ctx context.Context, projectID string) error

	// ListProjects retrieves projects with pagination and filtering
	ListProjects(ctx context.Context, opts ListProjectsOptions) ([]*ProjectRecord, string, error)

	// ProjectExists checks if a project exists by ID
	ProjectExists(ctx context.Context, projectID string) (bool, error)
}

// ListProjectsOptions contains options for listing projects
type ListProjectsOptions struct {
	// Token for pagination
	NextToken *string
	// Maximum number of results to return
	MaxResults *int
	// Tag filter - projects must match all provided tags
	TagFilter map[string]string
}
