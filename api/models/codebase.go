// Package models provides data transfer objects and request/response models for the API
package models

import (
	"time"
)

// Provider represents the supported code repository providers
type Provider string

// Supported code repository providers
const (
	ProviderGitHub    Provider = "github"    // GitHub repository provider
	ProviderGitLab    Provider = "gitlab"    // GitLab repository provider
	ProviderBitbucket Provider = "bitbucket" // Bitbucket repository provider
	ProviderCustom    Provider = "custom"    // Custom repository provider
)

// IsValid checks if the provider is one of the supported values
func (p Provider) IsValid() bool {
	switch p {
	case ProviderGitHub, ProviderGitLab, ProviderBitbucket, ProviderCustom:
		return true
	default:
		return false
	}
}

// String returns the string representation of the provider
func (p Provider) String() string {
	return string(p)
}

// Codebase represents a Git-based repository attached to a Project
type Codebase struct {
	CodebaseID    string            `json:"codebaseId" db:"codebase_id"`
	ProjectID     string            `json:"projectId" db:"project_id"`
	Name          string            `json:"name" db:"name"`
	Provider      Provider          `json:"provider" db:"provider"`
	URL           string            `json:"url" db:"url"`
	DefaultBranch string            `json:"defaultBranch" db:"default_branch"`
	CreatedAt     time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time         `json:"updatedAt" db:"updated_at"`
	Metadata      map[string]string `json:"metadata,omitempty" db:"metadata"`
	Tags          map[string]string `json:"tags,omitempty" db:"tags"`
}

// CreateCodebaseRequest represents the request to create a new codebase
type CreateCodebaseRequest struct {
	ProjectID     string            `json:"projectId" validate:"required,project_id" uri:"project_id"`
	Name          string            `json:"name" validate:"required,min=1,max=255"`
	Provider      Provider          `json:"provider" validate:"required,provider"`
	URL           string            `json:"url" validate:"required,url,max=2048"`
	DefaultBranch string            `json:"defaultBranch" validate:"required,min=1,max=255"`
	Tags          map[string]string `json:"tags,omitempty" validate:"omitempty,dive,keys,min=1,max=64,endkeys,min=1,max=255"`
}

// CreateCodebaseResponse represents the response after creating a codebase
type CreateCodebaseResponse struct {
	CodebaseID string `json:"codebaseId"`
	CreatedAt  string `json:"createdAt"`
}

// GetCodebaseRequest represents the request to get a codebase by ID
type GetCodebaseRequest struct {
	CodebaseID string `json:"codebaseId" validate:"required,uuid" uri:"id"`
}

// GetCodebaseResponse represents the response when retrieving a codebase
type GetCodebaseResponse struct {
	CodebaseID    string            `json:"codebaseId"`
	ProjectID     string            `json:"projectId"`
	Name          string            `json:"name"`
	Provider      Provider          `json:"provider"`
	URL           string            `json:"url"`
	DefaultBranch string            `json:"defaultBranch"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// UpdateCodebaseRequest represents the request to update a codebase
type UpdateCodebaseRequest struct {
	CodebaseID    string            `json:"codebaseId" validate:"required,uuid" uri:"id"`
	Name          *string           `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	DefaultBranch *string           `json:"defaultBranch,omitempty" validate:"omitempty,min=1,max=255"`
	Tags          map[string]string `json:"tags,omitempty" validate:"omitempty,dive,keys,min=1,max=64,endkeys,min=1,max=255"`
	Metadata      map[string]string `json:"metadata,omitempty" validate:"omitempty,dive,keys,min=1,max=64,endkeys,min=1,max=255"`
}

// UpdateCodebaseResponse represents the response after updating a codebase
type UpdateCodebaseResponse struct {
	CodebaseID string `json:"codebaseId"`
	UpdatedAt  string `json:"updatedAt"`
}

// DeleteCodebaseRequest represents the request to delete a codebase
type DeleteCodebaseRequest struct {
	CodebaseID string `json:"codebaseId" validate:"required,uuid" uri:"id"`
}

// DeleteCodebaseResponse represents the response after deleting a codebase
type DeleteCodebaseResponse struct {
	Success bool `json:"success"`
}

// ListCodebasesRequest represents the request to list codebases
type ListCodebasesRequest struct {
	ProjectID  *string `json:"projectId,omitempty" validate:"omitempty,project_id" form:"project_id"`
	TagFilter  *string `json:"tagFilter,omitempty" validate:"omitempty,tag_filter" form:"tag_filter"`
	NextToken  *string `json:"nextToken,omitempty" validate:"omitempty,max=1024" form:"next_token"`
	MaxResults *int    `json:"maxResults,omitempty" validate:"omitempty,min=1,max=100" form:"max_results"`
}

// CodebaseSummary represents a summary of a codebase for listing purposes
type CodebaseSummary struct {
	CodebaseID    string            `json:"codebaseId"`
	ProjectID     string            `json:"projectId"`
	Name          string            `json:"name"`
	Provider      Provider          `json:"provider"`
	URL           string            `json:"url"`
	DefaultBranch string            `json:"defaultBranch"`
	CreatedAt     string            `json:"createdAt"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// ListCodebasesResponse represents the response when listing codebases
type ListCodebasesResponse struct {
	Codebases []CodebaseSummary `json:"codebases"`
	NextToken *string           `json:"nextToken,omitempty"`
}
