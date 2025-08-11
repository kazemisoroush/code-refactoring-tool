// Package repository provides data access layer for codebase configurations
package repository

import (
	"context"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// CodebaseConfigRecord represents the codebase configuration data stored in the database
type CodebaseConfigRecord struct {
	ConfigID    string                   `json:"config_id" db:"config_id"`
	Name        string                   `json:"name" db:"name"`
	Description *string                  `json:"description,omitempty" db:"description"`
	Provider    string                   `json:"provider" db:"provider"`
	URL         string                   `json:"url" db:"url"`
	CreatedAt   time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at" db:"updated_at"`
	Tags        map[string]string        `json:"tags,omitempty" db:"tags"`
	Config      models.GitProviderConfig `json:"config" db:"config"`
}

// ToGetCodebaseConfigResponse converts CodebaseConfigRecord to GetCodebaseConfigResponse
func (r *CodebaseConfigRecord) ToGetCodebaseConfigResponse() *models.GetCodebaseConfigResponse {
	return &models.GetCodebaseConfigResponse{
		ConfigID:    r.ConfigID,
		Name:        r.Name,
		Description: r.Description,
		Provider:    models.Provider(r.Provider),
		URL:         r.URL,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
		Tags:        r.Tags,
		Config:      r.redactSensitiveConfig(),
	}
}

// ToCodebaseConfigSummary converts CodebaseConfigRecord to CodebaseConfigSummary
func (r *CodebaseConfigRecord) ToCodebaseConfigSummary() models.CodebaseConfigSummary {
	return models.CodebaseConfigSummary{
		ConfigID:  r.ConfigID,
		Name:      r.Name,
		Provider:  models.Provider(r.Provider),
		URL:       r.URL,
		CreatedAt: r.CreatedAt.UTC().Format(time.RFC3339),
		Tags:      r.Tags,
	}
}

// redactSensitiveConfig creates a redacted version of the configuration for safe API responses
func (r *CodebaseConfigRecord) redactSensitiveConfig() models.GitProviderConfigRedacted {
	redacted := models.GitProviderConfigRedacted{
		AuthType: r.Config.AuthType,
	}

	if r.Config.GitHub != nil {
		redacted.GitHub = &models.GitHubConfigRedacted{
			Organization:  r.Config.GitHub.Organization,
			Repository:    r.Config.GitHub.Repository,
			Owner:         r.Config.GitHub.Owner,
			DefaultBranch: r.Config.GitHub.DefaultBranch,
			HasToken:      r.Config.GitHub.Token != "",
		}
	}

	if r.Config.GitLab != nil {
		redacted.GitLab = &models.GitLabConfigRedacted{
			BaseURL:       r.Config.GitLab.BaseURL,
			ProjectID:     r.Config.GitLab.ProjectID,
			Namespace:     r.Config.GitLab.Namespace,
			DefaultBranch: r.Config.GitLab.DefaultBranch,
			HasToken:      r.Config.GitLab.Token != "",
		}
	}

	if r.Config.Bitbucket != nil {
		redacted.Bitbucket = &models.BitbucketConfigRedacted{
			Username:       r.Config.Bitbucket.Username,
			Workspace:      r.Config.Bitbucket.Workspace,
			Repository:     r.Config.Bitbucket.Repository,
			DefaultBranch:  r.Config.Bitbucket.DefaultBranch,
			HasAppPassword: r.Config.Bitbucket.AppPassword != "",
		}
	}

	if r.Config.Custom != nil {
		redacted.Custom = &models.CustomGitConfigRedacted{
			BaseURL:       r.Config.Custom.BaseURL,
			Headers:       r.Config.Custom.Headers,
			DefaultBranch: r.Config.Custom.DefaultBranch,
			HasToken:      r.Config.Custom.Token != "",
			HasUsername:   r.Config.Custom.Username != "",
			HasPassword:   r.Config.Custom.Password != "",
			HasSSHKey:     r.Config.Custom.SSHKey != "",
		}
	}

	return redacted
}

// CodebaseConfigRepository defines the interface for codebase configuration data operations
//
//go:generate mockgen -destination=./mocks/mock_codebase_config_repository.go -mock_names=CodebaseConfigRepository=MockCodebaseConfigRepository -package=mocks . CodebaseConfigRepository
type CodebaseConfigRepository interface {
	// CreateCodebaseConfig creates a new codebase configuration record
	CreateCodebaseConfig(ctx context.Context, config *CodebaseConfigRecord) error

	// GetCodebaseConfig retrieves a codebase configuration by ID
	GetCodebaseConfig(ctx context.Context, configID string) (*CodebaseConfigRecord, error)

	// UpdateCodebaseConfig updates an existing codebase configuration record
	UpdateCodebaseConfig(ctx context.Context, config *CodebaseConfigRecord) error

	// DeleteCodebaseConfig deletes a codebase configuration by ID
	DeleteCodebaseConfig(ctx context.Context, configID string) error

	// ListCodebaseConfigs retrieves codebase configurations with pagination and filtering
	ListCodebaseConfigs(ctx context.Context, opts ListCodebaseConfigsOptions) ([]*CodebaseConfigRecord, string, error)

	// CodebaseConfigExists checks if a codebase configuration exists by ID
	CodebaseConfigExists(ctx context.Context, configID string) (bool, error)
}

// ListCodebaseConfigsOptions contains options for listing codebase configurations
type ListCodebaseConfigsOptions struct {
	// Token for pagination
	NextToken *string
	// Maximum number of results to return
	MaxResults *int
	// Filter by provider
	ProviderFilter *models.Provider
	// Tag filter - configurations must match all provided tags
	TagFilter map[string]string
}
