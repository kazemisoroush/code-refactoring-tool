// Package services provides business logic for codebase configuration operations
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
)

// CodebaseConfigService defines the interface for codebase configuration business logic
//
//go:generate mockgen -destination=./mocks/mock_codebase_config_service.go -mock_names=CodebaseConfigService=MockCodebaseConfigService -package=mocks . CodebaseConfigService
type CodebaseConfigService interface {
	// CreateCodebaseConfig creates a new codebase configuration
	CreateCodebaseConfig(ctx context.Context, request models.CreateCodebaseConfigRequest) (*models.CreateCodebaseConfigResponse, error)

	// GetCodebaseConfig retrieves a codebase configuration by ID
	GetCodebaseConfig(ctx context.Context, configID string) (*models.GetCodebaseConfigResponse, error)

	// UpdateCodebaseConfig updates an existing codebase configuration
	UpdateCodebaseConfig(ctx context.Context, request models.UpdateCodebaseConfigRequest) (*models.UpdateCodebaseConfigResponse, error)

	// DeleteCodebaseConfig deletes a codebase configuration by ID
	DeleteCodebaseConfig(ctx context.Context, configID string) (*models.DeleteCodebaseConfigResponse, error)

	// ListCodebaseConfigs retrieves codebase configurations with pagination and filtering
	ListCodebaseConfigs(ctx context.Context, request models.ListCodebaseConfigsRequest) (*models.ListCodebaseConfigsResponse, error)
}

// DefaultCodebaseConfigService implements CodebaseConfigService
type DefaultCodebaseConfigService struct {
	repository repository.CodebaseConfigRepository
}

// NewDefaultCodebaseConfigService creates a new DefaultCodebaseConfigService
func NewDefaultCodebaseConfigService(
	repository repository.CodebaseConfigRepository,
) CodebaseConfigService {
	return &DefaultCodebaseConfigService{
		repository: repository,
	}
}

// CreateCodebaseConfig creates a new codebase configuration
func (s *DefaultCodebaseConfigService) CreateCodebaseConfig(ctx context.Context, request models.CreateCodebaseConfigRequest) (*models.CreateCodebaseConfigResponse, error) {
	// Validate the provider
	if !request.Provider.IsValid() {
		return nil, fmt.Errorf("invalid provider: %s", request.Provider)
	}

	// Validate provider-specific configuration
	if err := s.validateProviderConfig(request.Provider, request.Config); err != nil {
		return nil, fmt.Errorf("invalid provider configuration: %w", err)
	}

	// Generate unique configuration ID
	configID := generateCodebaseConfigID()

	now := time.Now().UTC()
	record := &repository.CodebaseConfigRecord{
		ConfigID:      configID,
		Name:          request.Name,
		Description:   request.Description,
		Provider:      string(request.Provider),
		URL:           request.URL,
		DefaultBranch: request.DefaultBranch,
		Status:        string(models.CodebaseConfigStatusActive),
		CreatedAt:     now,
		UpdatedAt:     now,
		Tags:          request.Tags,
		Metadata:      request.Metadata,
		Config:        request.Config,
	}

	// Create the configuration in the repository
	if err := s.repository.CreateCodebaseConfig(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create codebase configuration: %w", err)
	}

	return &models.CreateCodebaseConfigResponse{
		ConfigID:  configID,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}

// GetCodebaseConfig retrieves a codebase configuration by ID
func (s *DefaultCodebaseConfigService) GetCodebaseConfig(ctx context.Context, configID string) (*models.GetCodebaseConfigResponse, error) {
	record, err := s.repository.GetCodebaseConfig(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase configuration: %w", err)
	}

	return record.ToGetCodebaseConfigResponse(), nil
}

// UpdateCodebaseConfig updates an existing codebase configuration
func (s *DefaultCodebaseConfigService) UpdateCodebaseConfig(ctx context.Context, request models.UpdateCodebaseConfigRequest) (*models.UpdateCodebaseConfigResponse, error) {
	// Get existing configuration
	existing, err := s.repository.GetCodebaseConfig(ctx, request.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing codebase configuration: %w", err)
	}

	// Update only provided fields
	now := time.Now().UTC()

	if request.Name != nil {
		existing.Name = *request.Name
	}
	if request.Description != nil {
		existing.Description = request.Description
	}
	if request.URL != nil {
		existing.URL = *request.URL
	}
	if request.DefaultBranch != nil {
		existing.DefaultBranch = *request.DefaultBranch
	}
	if request.Config != nil {
		// Validate provider-specific configuration
		provider := models.Provider(existing.Provider)
		if err := s.validateProviderConfig(provider, *request.Config); err != nil {
			return nil, fmt.Errorf("invalid provider configuration: %w", err)
		}
		existing.Config = *request.Config
	}
	if request.Tags != nil {
		existing.Tags = request.Tags
	}
	if request.Metadata != nil {
		existing.Metadata = request.Metadata
	}

	existing.UpdatedAt = now

	// Update in repository
	if err := s.repository.UpdateCodebaseConfig(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update codebase configuration: %w", err)
	}

	return &models.UpdateCodebaseConfigResponse{
		ConfigID:  request.ConfigID,
		UpdatedAt: now.Format(time.RFC3339),
	}, nil
}

// DeleteCodebaseConfig deletes a codebase configuration by ID
func (s *DefaultCodebaseConfigService) DeleteCodebaseConfig(ctx context.Context, configID string) (*models.DeleteCodebaseConfigResponse, error) {
	// Check if configuration exists
	exists, err := s.repository.CodebaseConfigExists(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if codebase configuration exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("codebase configuration not found")
	}

	// Delete the configuration
	if err := s.repository.DeleteCodebaseConfig(ctx, configID); err != nil {
		return nil, fmt.Errorf("failed to delete codebase configuration: %w", err)
	}

	return &models.DeleteCodebaseConfigResponse{
		Success: true,
	}, nil
}

// ListCodebaseConfigs retrieves codebase configurations with pagination and filtering
func (s *DefaultCodebaseConfigService) ListCodebaseConfigs(ctx context.Context, request models.ListCodebaseConfigsRequest) (*models.ListCodebaseConfigsResponse, error) {
	opts := repository.ListCodebaseConfigsOptions{
		NextToken:      request.NextToken,
		MaxResults:     request.MaxResults,
		ProviderFilter: request.ProviderFilter,
		TagFilter:      request.TagFilter,
	}

	records, nextToken, err := s.repository.ListCodebaseConfigs(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list codebase configurations: %w", err)
	}

	// Convert records to summaries
	summaries := make([]models.CodebaseConfigSummary, len(records))
	for i, record := range records {
		summaries[i] = record.ToCodebaseConfigSummary()
	}

	response := &models.ListCodebaseConfigsResponse{
		Configs: summaries,
	}

	if nextToken != "" {
		response.NextToken = &nextToken
	}

	return response, nil
}

// validateProviderConfig validates provider-specific configuration
func (s *DefaultCodebaseConfigService) validateProviderConfig(provider models.Provider, config models.GitProviderConfig) error {
	// Validate provider-specific fields
	if err := s.validateProviderFields(provider, config); err != nil {
		return err
	}

	// Validate authentication configuration
	return s.validateAuthConfig(provider, config)
}

// validateProviderFields validates provider-specific required fields
func (s *DefaultCodebaseConfigService) validateProviderFields(provider models.Provider, config models.GitProviderConfig) error {
	switch provider {
	case models.ProviderGitHub:
		return s.validateGitHubConfig(config)
	case models.ProviderGitLab:
		return s.validateGitLabConfig(config)
	case models.ProviderBitbucket:
		return s.validateBitbucketConfig(config)
	case models.ProviderCustom:
		return s.validateCustomConfig(config)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// validateGitHubConfig validates GitHub-specific configuration
func (s *DefaultCodebaseConfigService) validateGitHubConfig(config models.GitProviderConfig) error {
	if config.GitHub == nil {
		return fmt.Errorf("GitHub configuration is required for GitHub provider")
	}
	if config.GitHub.Owner == "" {
		return fmt.Errorf("GitHub owner is required")
	}
	if config.GitHub.Repository == "" {
		return fmt.Errorf("GitHub repository is required")
	}
	return nil
}

// validateGitLabConfig validates GitLab-specific configuration
func (s *DefaultCodebaseConfigService) validateGitLabConfig(config models.GitProviderConfig) error {
	if config.GitLab == nil {
		return fmt.Errorf("GitLab configuration is required for GitLab provider")
	}
	if config.GitLab.ProjectID == "" {
		return fmt.Errorf("GitLab project ID is required")
	}
	if config.GitLab.Namespace == "" {
		return fmt.Errorf("GitLab namespace is required")
	}
	return nil
}

// validateBitbucketConfig validates Bitbucket-specific configuration
func (s *DefaultCodebaseConfigService) validateBitbucketConfig(config models.GitProviderConfig) error {
	if config.Bitbucket == nil {
		return fmt.Errorf("bitbucket configuration is required for Bitbucket provider")
	}
	if config.Bitbucket.Workspace == "" {
		return fmt.Errorf("bitbucket workspace is required")
	}
	if config.Bitbucket.Repository == "" {
		return fmt.Errorf("bitbucket repository is required")
	}
	return nil
}

// validateCustomConfig validates custom provider configuration
func (s *DefaultCodebaseConfigService) validateCustomConfig(config models.GitProviderConfig) error {
	if config.Custom == nil {
		return fmt.Errorf("custom configuration is required for custom provider")
	}
	if config.Custom.BaseURL == "" {
		return fmt.Errorf("custom base URL is required")
	}
	return nil
}

// validateAuthConfig validates authentication configuration
func (s *DefaultCodebaseConfigService) validateAuthConfig(provider models.Provider, config models.GitProviderConfig) error {
	switch config.AuthType {
	case models.GitAuthTypeToken:
		return s.validateTokenAuth(provider, config)
	case models.GitAuthTypeBasic:
		return s.validateBasicAuth(provider, config)
	case models.GitAuthTypeSSH:
		return s.validateSSHAuth(provider, config)
	case models.GitAuthTypeOAuth:
		return fmt.Errorf("OAuth authentication is not yet supported")
	default:
		return fmt.Errorf("unsupported authentication type: %s", config.AuthType)
	}
}

// validateTokenAuth validates token-based authentication
func (s *DefaultCodebaseConfigService) validateTokenAuth(provider models.Provider, config models.GitProviderConfig) error {
	switch provider {
	case models.ProviderGitHub:
		if config.GitHub.Token == "" {
			return fmt.Errorf("token is required for GitHub token authentication")
		}
	case models.ProviderGitLab:
		if config.GitLab.Token == "" {
			return fmt.Errorf("token is required for GitLab token authentication")
		}
	case models.ProviderBitbucket:
		if config.Bitbucket.AppPassword == "" {
			return fmt.Errorf("app password is required for Bitbucket token authentication")
		}
	case models.ProviderCustom:
		if config.Custom.Token == "" {
			return fmt.Errorf("token is required for custom token authentication")
		}
	}
	return nil
}

// validateBasicAuth validates basic authentication
func (s *DefaultCodebaseConfigService) validateBasicAuth(provider models.Provider, config models.GitProviderConfig) error {
	if provider != models.ProviderCustom {
		return fmt.Errorf("basic authentication is only supported for custom providers")
	}
	if config.Custom.Username == "" || config.Custom.Password == "" {
		return fmt.Errorf("username and password are required for basic authentication")
	}
	return nil
}

// validateSSHAuth validates SSH key authentication
func (s *DefaultCodebaseConfigService) validateSSHAuth(provider models.Provider, config models.GitProviderConfig) error {
	if provider != models.ProviderCustom {
		return fmt.Errorf("SSH authentication is only supported for custom providers")
	}
	if config.Custom.SSHKey == "" {
		return fmt.Errorf("SSH key is required for SSH authentication")
	}
	return nil
}

// generateCodebaseConfigID generates a unique codebase configuration ID
func generateCodebaseConfigID() string {
	// Generate UUID and format as AWS-style resource identifier
	id := uuid.New().String()
	return fmt.Sprintf("config-%s", id[:13])
}
