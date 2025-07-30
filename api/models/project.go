// Package models provides data structures for API requests and responses
package models

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	// Human-readable project name
	Name string `json:"name" validate:"required,min=1,max=100" example:"my-project"`
	// Optional project summary
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"A sample project for code analysis"`
	// Optional programming language
	Language *string `json:"language,omitempty" validate:"omitempty,oneof=go javascript typescript python java csharp rust cpp c ruby php kotlin swift scala other" example:"go"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod,team:backend"`
} //@name CreateProjectRequest

// CreateProjectResponse represents the response when creating a project
type CreateProjectResponse struct {
	// Unique identifier for the project
	ProjectID string `json:"project_id" example:"proj-12345-abcde"`
	// Timestamp when the project was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
} //@name CreateProjectResponse

// GetProjectRequest represents the request to get a project
type GetProjectRequest struct {
	// Unique identifier for the project
	ProjectID string `uri:"id" validate:"required,project_id" example:"proj-12345-abcde"`
} //@name GetProjectRequest

// GetProjectResponse represents the response when getting a project
type GetProjectResponse struct {
	// Unique identifier for the project
	ProjectID string `json:"project_id" example:"proj-12345-abcde"`
	// Human-readable project name
	Name string `json:"name" example:"my-project"`
	// Optional project summary
	Description *string `json:"description,omitempty" example:"A sample project for code analysis"`
	// Optional programming language
	Language *string `json:"language,omitempty" example:"go"`
	// Timestamp when the project was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when the project was last updated
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" example:"env:prod,team:backend"`
	// Optional metadata
	Metadata map[string]string `json:"metadata,omitempty" example:"version:1.0.0"`
} //@name GetProjectResponse

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	// Unique identifier for the project
	ProjectID string `uri:"id" validate:"required,project_id" example:"proj-12345-abcde"`
	// Optional human-readable project name
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"updated-project"`
	// Optional project summary
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated project description"`
	// Optional programming language
	Language *string `json:"language,omitempty" validate:"omitempty,oneof=go javascript typescript python java csharp rust cpp c ruby php kotlin swift scala other" example:"python"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:staging,team:frontend"`
	// Optional metadata
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty,max=20,dive,keys,min=1,max=100,endkeys,min=1,max=500" example:"version:1.1.0"`
} //@name UpdateProjectRequest

// UpdateProjectResponse represents the response when updating a project
type UpdateProjectResponse struct {
	// Unique identifier for the project
	ProjectID string `json:"project_id" example:"proj-12345-abcde"`
	// Timestamp when the project was last updated
	UpdatedAt string `json:"updated_at" example:"2024-01-15T11:30:00Z"`
} //@name UpdateProjectResponse

// DeleteProjectRequest represents the request to delete a project
type DeleteProjectRequest struct {
	// Unique identifier for the project
	ProjectID string `uri:"id" validate:"required,project_id" example:"proj-12345-abcde"`
} //@name DeleteProjectRequest

// DeleteProjectResponse represents the response when deleting a project
type DeleteProjectResponse struct {
	// Indicates whether the deletion was successful
	Success bool `json:"success" example:"true"`
} //@name DeleteProjectResponse

// ListProjectsRequest represents the request to list projects
type ListProjectsRequest struct {
	// Token for pagination
	NextToken *string `form:"next_token,omitempty" validate:"omitempty,min=1" example:"eyJpZCI6InByb2otMTIzNDUifQ=="`
	// Maximum number of results to return
	MaxResults *int `form:"max_results,omitempty" validate:"omitempty,min=1,max=100" example:"50"`
	// Optional tag filter - projects must match all provided tags
	TagFilter map[string]string `form:"tag_filter,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod"`
} //@name ListProjectsRequest

// ListProjectsResponse represents the response when listing projects
type ListProjectsResponse struct {
	// List of project summaries
	Projects []ProjectSummary `json:"projects"`
	// Token for next page of results
	NextToken *string `json:"next_token,omitempty" example:"eyJpZCI6InByb2otNjc4OTAifQ=="`
} //@name ListProjectsResponse

// ProjectSummary represents a summary of a project for listing
type ProjectSummary struct {
	// Unique identifier for the project
	ProjectID string `json:"project_id" example:"proj-12345-abcde"`
	// Human-readable project name
	Name string `json:"name" example:"my-project"`
	// Timestamp when the project was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" example:"env:prod,team:backend"`
} //@name ProjectSummary

// ProjectStatus represents the status of a project
type ProjectStatus string

// Project status constants
const (
	// ProjectStatusActive indicates the project is active
	ProjectStatusActive ProjectStatus = "active"
	// ProjectStatusArchived indicates the project is archived
	ProjectStatusArchived ProjectStatus = "archived"
	// ProjectStatusDeleted indicates the project has been deleted
	ProjectStatusDeleted ProjectStatus = "deleted"
)
