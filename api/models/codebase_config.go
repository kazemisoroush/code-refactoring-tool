// Package models provides data structures for codebase configuration API requests and responses
package models

import (
	"time"
)

// CodebaseConfigStatus represents the status of a codebase configuration
type CodebaseConfigStatus string

// Codebase configuration status constants
const (
	// CodebaseConfigStatusActive indicates the configuration is active
	CodebaseConfigStatusActive CodebaseConfigStatus = "active"
	// CodebaseConfigStatusInactive indicates the configuration is inactive
	CodebaseConfigStatusInactive CodebaseConfigStatus = "inactive"
	// CodebaseConfigStatusDeleted indicates the configuration has been deleted
	CodebaseConfigStatusDeleted CodebaseConfigStatus = "deleted"
)

// CodebaseConfig represents a codebase configuration profile that can be reused across projects
type CodebaseConfig struct {
	ConfigID      string               `json:"config_id" db:"config_id"`
	Name          string               `json:"name" db:"name"`
	Description   *string              `json:"description,omitempty" db:"description"`
	Provider      Provider             `json:"provider" db:"provider"`
	URL           string               `json:"url" db:"url"`
	DefaultBranch string               `json:"default_branch" db:"default_branch"`
	Status        CodebaseConfigStatus `json:"status" db:"status"`
	CreatedAt     time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at" db:"updated_at"`
	Tags          map[string]string    `json:"tags,omitempty" db:"tags"`
	Metadata      map[string]string    `json:"metadata,omitempty" db:"metadata"`

	// Git provider configuration (contains sensitive data)
	Config GitProviderConfig `json:"config" db:"config"`
}

// CreateCodebaseConfigRequest represents the request to create a new codebase configuration
type CreateCodebaseConfigRequest struct {
	// Human-readable configuration name
	Name string `json:"name" validate:"required,min=1,max=100" example:"my-github-config"`
	// Optional configuration description
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"GitHub configuration for production repositories"`
	// Repository provider type
	Provider Provider `json:"provider" validate:"required,provider" example:"github"`
	// Repository URL
	URL string `json:"url" validate:"required,url,max=2048" example:"https://github.com/owner/repo.git"`
	// Default branch to use
	DefaultBranch string `json:"default_branch" validate:"required,min=1,max=255" example:"main"`
	// Provider-specific configuration
	Config GitProviderConfig `json:"config" validate:"required"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod,team:backend"`
	// Optional metadata
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty,max=20,dive,keys,min=1,max=100,endkeys,min=1,max=500" example:"owner:team-lead"`
} //@name CreateCodebaseConfigRequest

// CreateCodebaseConfigResponse represents the response when creating a codebase configuration
type CreateCodebaseConfigResponse struct {
	// Unique identifier for the configuration
	ConfigID string `json:"config_id" example:"config-12345-abcde"`
	// Timestamp when the configuration was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
} //@name CreateCodebaseConfigResponse

// GetCodebaseConfigRequest represents the request to get a codebase configuration
type GetCodebaseConfigRequest struct {
	// Unique identifier for the configuration
	ConfigID string `uri:"config_id" validate:"required,config_id" example:"config-12345-abcde"`
} //@name GetCodebaseConfigRequest

// GetCodebaseConfigResponse represents the response when getting a codebase configuration
type GetCodebaseConfigResponse struct {
	// Unique identifier for the configuration
	ConfigID string `json:"config_id" example:"config-12345-abcde"`
	// Human-readable configuration name
	Name string `json:"name" example:"my-github-config"`
	// Optional configuration description
	Description *string `json:"description,omitempty" example:"GitHub configuration for production repositories"`
	// Repository provider type
	Provider Provider `json:"provider" example:"github"`
	// Repository URL
	URL string `json:"url" example:"https://github.com/owner/repo.git"`
	// Default branch to use
	DefaultBranch string `json:"default_branch" example:"main"`
	// Configuration status
	Status CodebaseConfigStatus `json:"status" example:"active"`
	// Timestamp when the configuration was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when the configuration was last updated
	UpdatedAt string `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" example:"env:prod,team:backend"`
	// Optional metadata
	Metadata map[string]string `json:"metadata,omitempty" example:"owner:team-lead"`
	// Provider-specific configuration (sensitive data redacted)
	Config GitProviderConfigRedacted `json:"config"`
} //@name GetCodebaseConfigResponse

// GitProviderConfigRedacted represents provider-specific configuration with sensitive data redacted
type GitProviderConfigRedacted struct {
	// Authentication method
	AuthType GitAuthType `json:"auth_type"`
	// For GitHub (with sensitive fields redacted)
	GitHub *GitHubConfigRedacted `json:"github,omitempty"`
	// For GitLab (with sensitive fields redacted)
	GitLab *GitLabConfigRedacted `json:"gitlab,omitempty"`
	// For Bitbucket (with sensitive fields redacted)
	Bitbucket *BitbucketConfigRedacted `json:"bitbucket,omitempty"`
	// For custom Git providers (with sensitive fields redacted)
	Custom *CustomGitConfigRedacted `json:"custom,omitempty"`
}

// GitHubConfigRedacted represents GitHub-specific configuration with sensitive data redacted
type GitHubConfigRedacted struct {
	Organization string `json:"organization,omitempty"`
	Repository   string `json:"repository"`
	Owner        string `json:"owner"`
	// Token is redacted for security
	HasToken bool `json:"has_token"`
}

// GitLabConfigRedacted represents GitLab-specific configuration with sensitive data redacted
type GitLabConfigRedacted struct {
	BaseURL   string `json:"base_url,omitempty"`
	ProjectID string `json:"project_id"`
	Namespace string `json:"namespace"`
	// Token is redacted for security
	HasToken bool `json:"has_token"`
}

// BitbucketConfigRedacted represents Bitbucket-specific configuration with sensitive data redacted
type BitbucketConfigRedacted struct {
	Username   string `json:"username,omitempty"`
	Workspace  string `json:"workspace"`
	Repository string `json:"repository"`
	// AppPassword is redacted for security
	HasAppPassword bool `json:"has_app_password"`
}

// CustomGitConfigRedacted represents configuration for custom Git providers with sensitive data redacted
type CustomGitConfigRedacted struct {
	BaseURL string            `json:"base_url"`
	Headers map[string]string `json:"headers,omitempty"`
	// Sensitive fields are redacted for security
	HasToken    bool `json:"has_token"`
	HasUsername bool `json:"has_username"`
	HasPassword bool `json:"has_password"`
	HasSSHKey   bool `json:"has_ssh_key"`
}

// UpdateCodebaseConfigRequest represents the request to update a codebase configuration
type UpdateCodebaseConfigRequest struct {
	// Unique identifier for the configuration
	ConfigID string `uri:"config_id" validate:"required,config_id" example:"config-12345-abcde"`
	// Optional human-readable configuration name
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"updated-github-config"`
	// Optional configuration description
	Description *string `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated GitHub configuration"`
	// Optional repository URL
	URL *string `json:"url,omitempty" validate:"omitempty,url,max=2048" example:"https://github.com/owner/new-repo.git"`
	// Optional default branch to use
	DefaultBranch *string `json:"default_branch,omitempty" validate:"omitempty,min=1,max=255" example:"develop"`
	// Optional provider-specific configuration
	Config *GitProviderConfig `json:"config,omitempty"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:staging,team:frontend"`
	// Optional metadata
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty,max=20,dive,keys,min=1,max=100,endkeys,min=1,max=500" example:"owner:new-team-lead"`
} //@name UpdateCodebaseConfigRequest

// UpdateCodebaseConfigResponse represents the response when updating a codebase configuration
type UpdateCodebaseConfigResponse struct {
	// Unique identifier for the configuration
	ConfigID string `json:"config_id" example:"config-12345-abcde"`
	// Timestamp when the configuration was last updated
	UpdatedAt string `json:"updated_at" example:"2024-01-15T11:30:00Z"`
} //@name UpdateCodebaseConfigResponse

// DeleteCodebaseConfigRequest represents the request to delete a codebase configuration
type DeleteCodebaseConfigRequest struct {
	// Unique identifier for the configuration
	ConfigID string `uri:"config_id" validate:"required,config_id" example:"config-12345-abcde"`
} //@name DeleteCodebaseConfigRequest

// DeleteCodebaseConfigResponse represents the response when deleting a codebase configuration
type DeleteCodebaseConfigResponse struct {
	// Indicates whether the deletion was successful
	Success bool `json:"success" example:"true"`
} //@name DeleteCodebaseConfigResponse

// ListCodebaseConfigsRequest represents the request to list codebase configurations
type ListCodebaseConfigsRequest struct {
	// Token for pagination
	NextToken *string `form:"next_token,omitempty" validate:"omitempty,min=1" example:"eyJpZCI6ImNvbmZpZy0xMjM0NSJ9"`
	// Maximum number of results to return
	MaxResults *int `form:"max_results,omitempty" validate:"omitempty,min=1,max=100" example:"50"`
	// Filter by provider
	ProviderFilter *Provider `form:"provider_filter,omitempty" validate:"omitempty,provider" example:"github"`
	// Optional tag filter - configurations must match all provided tags
	TagFilter map[string]string `form:"tag_filter,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod"`
} //@name ListCodebaseConfigsRequest

// ListCodebaseConfigsResponse represents the response when listing codebase configurations
type ListCodebaseConfigsResponse struct {
	// List of configuration summaries
	Configs []CodebaseConfigSummary `json:"configs"`
	// Token for next page of results
	NextToken *string `json:"next_token,omitempty" example:"eyJpZCI6ImNvbmZpZy02Nzg5MCJ9"`
} //@name ListCodebaseConfigsResponse

// CodebaseConfigSummary represents a summary of a codebase configuration for listing
type CodebaseConfigSummary struct {
	// Unique identifier for the configuration
	ConfigID string `json:"config_id" example:"config-12345-abcde"`
	// Human-readable configuration name
	Name string `json:"name" example:"my-github-config"`
	// Repository provider type
	Provider Provider `json:"provider" example:"github"`
	// Repository URL
	URL string `json:"url" example:"https://github.com/owner/repo.git"`
	// Configuration status
	Status CodebaseConfigStatus `json:"status" example:"active"`
	// Timestamp when the configuration was created
	CreatedAt string `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Optional user-defined key-value tags
	Tags map[string]string `json:"tags,omitempty" example:"env:prod,team:backend"`
} //@name CodebaseConfigSummary
