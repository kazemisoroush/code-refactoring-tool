package factory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

// DefaultAIInfrastructureFactory implements the AIInfrastructureFactory interface
type DefaultAIInfrastructureFactory struct {
	// AWS configuration for Bedrock
	awsConfig aws.Config

	// Base AI configuration from program config
	aiConfig config.AIConfig

	// Git configuration
	gitConfig config.GitConfig
}

// NewAIInfrastructureFactory creates a new AI infrastructure factory
func NewAIInfrastructureFactory(awsConfig aws.Config, aiConfig config.AIConfig, gitConfig config.GitConfig) AIInfrastructureFactory {
	return &DefaultAIInfrastructureFactory{
		awsConfig: awsConfig,
		aiConfig:  aiConfig,
		gitConfig: gitConfig,
	}
}

// CreateAgentInfrastructure creates AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) CreateAgentInfrastructure(ctx context.Context, provider models.AIProvider) (*AIInfrastructureResult, error) {
	switch provider {
	case models.AIProviderBedrock:
		return f.createBedrockInfrastructure(ctx)
	case models.AIProviderLocal:
		return f.createLocalInfrastructure(ctx)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// Private helper methods for creating provider-specific infrastructure

func (f *DefaultAIInfrastructureFactory) createBedrockInfrastructure(ctx context.Context) (*AIInfrastructureResult, error) {
	config := f.aiConfig.Bedrock

	// Create repository instance
	repo := codebase.NewGitHubCodebase(f.gitConfig)

	// Create Bedrock dependencies
	dataStore := storage.NewS3DataStore(f.awsConfig, config.S3BucketName, repo.GetPath())
	storageImpl := storage.NewRDSPostgresStorage(f.awsConfig, "lambda-arn-placeholder") // TODO: Add Lambda ARN to config
	ragImpl := rag.NewBedrockRAG(f.awsConfig, repo.GetPath(), config.KnowledgeBaseServiceRoleARN, config.RDSPostgres)

	// Create Bedrock RAG builder
	ragBuilder := builder.NewBedrockRAGBuilder(
		repo.GetPath(),
		dataStore,
		storageImpl,
		ragImpl,
	)

	// Create Bedrock agent builder
	agentBuilder := builder.NewBedrockAgentBuilder(
		f.awsConfig,
		repo.GetPath(),
		config.AgentServiceRoleARN,
	)

	// Create and run workflow
	wf, err := workflow.NewCreateBedrockSetupWorkflow(repo, ragBuilder, agentBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bedrock setup workflow: %w", err)
	}

	err = wf.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run Bedrock setup workflow: %w", err)
	}

	// Get resource IDs from workflow
	setupWf := wf.(*workflow.CreateBedrockSetupWorkflow)
	vectorStoreID, ragID, agentID, agentVersion := setupWf.GetResourceIDs()

	return &AIInfrastructureResult{
		AgentID:         agentID,
		AgentVersion:    agentVersion,
		KnowledgeBaseID: ragID,
		VectorStoreID:   vectorStoreID,
		Status:          models.AgentStatusInitializing,
		Metadata: map[string]interface{}{
			"provider":     "bedrock",
			"region":       config.Region,
			"model":        config.FoundationModel,
			"service_role": config.AgentServiceRoleARN,
			"s3_bucket":    config.S3BucketName,
		},
	}, nil
}

func (f *DefaultAIInfrastructureFactory) createLocalInfrastructure(ctx context.Context) (*AIInfrastructureResult, error) {
	config := f.aiConfig.Local

	// Create repository instance
	repo := codebase.NewGitHubCodebase(f.gitConfig)

	// Create RAG builder
	ragBuilder := builder.NewLocalRAGBuilder(
		repo.GetPath(),
		config.ChromaURL,
		config.EmbeddingModel,
	)

	// Create agent builder
	agentBuilder := builder.NewLocalAgentBuilder(config.OllamaURL, config.Model)

	// Create and run workflow
	wf, err := workflow.NewCreateLocalSetupWorkflow(repo, ragBuilder, agentBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to create local setup workflow: %w", err)
	}

	err = wf.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to run local setup workflow: %w", err)
	}

	// Get resource IDs from workflow
	setupWf := wf.(*workflow.CreateLocalSetupWorkflow)
	vectorStoreID, ragID, agentID, agentVersion := setupWf.GetResourceIDs()

	return &AIInfrastructureResult{
		AgentID:         agentID,
		AgentVersion:    agentVersion,
		KnowledgeBaseID: ragID,
		VectorStoreID:   vectorStoreID,
		Status:          models.AgentStatusInitializing,
		Metadata: map[string]interface{}{
			"provider":   "local",
			"ollama_url": config.OllamaURL,
			"model":      config.Model,
		},
	}, nil
}

// ValidateAgentConfig validates an agent's AI provider configuration
func (f *DefaultAIInfrastructureFactory) ValidateAgentConfig(provider models.AIProvider) error {
	switch provider {
	case models.AIProviderBedrock:
		return f.validateBedrockConfig()
	case models.AIProviderLocal:
		return f.validateLocalConfig()
	default:
		return fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// Private validation methods
func (f *DefaultAIInfrastructureFactory) validateBedrockConfig() error {
	config := f.aiConfig.Bedrock
	if config.Region == "" {
		return fmt.Errorf("bedrock region is required")
	}
	if config.FoundationModel == "" {
		return fmt.Errorf("bedrock foundation model is required")
	}
	if config.AgentServiceRoleARN == "" {
		return fmt.Errorf("bedrock agent service role ARN is required")
	}
	return nil
}

func (f *DefaultAIInfrastructureFactory) validateLocalConfig() error {
	config := f.aiConfig.Local
	if config.OllamaURL == "" {
		return fmt.Errorf("ollama URL is required")
	}
	if config.Model == "" {
		return fmt.Errorf("local model is required")
	}
	return nil
}

// DestroyAgentInfrastructure cleans up AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) DestroyAgentInfrastructure(ctx context.Context, infrastructureID string) error {
	slog.Info("Destroying AI infrastructure", "infrastructure_id", infrastructureID)

	// In a real implementation, we would:
	// 1. Query the database to get the infrastructure metadata (provider, resource IDs, etc.)
	// 2. Create the appropriate teardown workflow based on the provider
	// 3. Run the teardown workflow

	// For now, we'll assume we can determine the provider from the infrastructureID format
	// This is a simplified implementation

	provider := f.determineProviderFromID(infrastructureID)

	switch provider {
	case models.AIProviderBedrock:
		return f.teardownBedrockInfrastructure(ctx, infrastructureID)
	case models.AIProviderLocal:
		return f.teardownLocalInfrastructure(ctx, infrastructureID)
	default:
		return fmt.Errorf("cannot determine provider for infrastructure ID: %s", infrastructureID)
	}
}

// determineProviderFromID attempts to determine the provider from the infrastructure ID
// In a real implementation, this would query a database
func (f *DefaultAIInfrastructureFactory) determineProviderFromID(infrastructureID string) models.AIProvider {
	// Simple heuristic based on ID format - in reality this would come from database
	if strings.Contains(infrastructureID, "bedrock") || strings.Contains(infrastructureID, "aws") {
		return models.AIProviderBedrock
	}
	if strings.Contains(infrastructureID, "local") || strings.Contains(infrastructureID, "ollama") {
		return models.AIProviderLocal
	}

	// Default to Bedrock if we can't determine
	return models.AIProviderBedrock
}

// teardownBedrockInfrastructure tears down Bedrock-specific infrastructure
func (f *DefaultAIInfrastructureFactory) teardownBedrockInfrastructure(ctx context.Context, infrastructureID string) error {
	config := f.aiConfig.Bedrock

	// Create repository instance (needed for cleanup)
	repo := codebase.NewGitHubCodebase(f.gitConfig)

	// Create Bedrock dependencies for teardown
	dataStore := storage.NewS3DataStore(f.awsConfig, config.S3BucketName, repo.GetPath())
	storageImpl := storage.NewRDSPostgresStorage(f.awsConfig, "lambda-arn-placeholder")
	ragImpl := rag.NewBedrockRAG(f.awsConfig, repo.GetPath(), config.KnowledgeBaseServiceRoleARN, config.RDSPostgres)

	// Create Bedrock builders for teardown
	ragBuilder := builder.NewBedrockRAGBuilder(repo.GetPath(), dataStore, storageImpl, ragImpl)
	agentBuilder := builder.NewBedrockAgentBuilder(f.awsConfig, repo.GetPath(), config.AgentServiceRoleARN)

	// Create teardown workflow with resource IDs
	// In a real implementation, these IDs would come from stored metadata
	wf, err := workflow.NewTeardownBedrockSetupWorkflowWithResources(
		repo,
		ragBuilder,
		agentBuilder,
		infrastructureID, // vectorStoreID
		infrastructureID, // ragID
		infrastructureID, // agentID
		"DRAFT",          // agentVersion - TODO: Get from metadata
	)
	if err != nil {
		return fmt.Errorf("failed to create Bedrock teardown workflow: %w", err)
	}

	err = wf.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run Bedrock teardown workflow: %w", err)
	}

	slog.Info("Bedrock infrastructure destroyed successfully", "infrastructure_id", infrastructureID)
	return nil
}

// teardownLocalInfrastructure tears down local-specific infrastructure
func (f *DefaultAIInfrastructureFactory) teardownLocalInfrastructure(ctx context.Context, infrastructureID string) error {
	config := f.aiConfig.Local

	// Create repository instance (needed for cleanup)
	repo := codebase.NewGitHubCodebase(f.gitConfig)

	// Create local builders for teardown
	ragBuilder := builder.NewLocalRAGBuilder(repo.GetPath(), config.ChromaURL, config.EmbeddingModel)
	agentBuilder := builder.NewLocalAgentBuilder(config.OllamaURL, config.Model)

	// Create teardown workflow with resource IDs
	// In a real implementation, these IDs would come from stored metadata
	wf, err := workflow.NewTeardownLocalSetupWorkflowWithResources(
		repo,
		ragBuilder,
		agentBuilder,
		infrastructureID, // vectorStoreID
		infrastructureID, // ragID
		infrastructureID, // agentID
		"v1",             // agentVersion - TODO: Get from metadata
	)
	if err != nil {
		return fmt.Errorf("failed to create local teardown workflow: %w", err)
	}

	err = wf.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run local teardown workflow: %w", err)
	}

	slog.Info("Local infrastructure destroyed successfully", "infrastructure_id", infrastructureID)
	return nil
}

// UpdateAgentInfrastructure updates existing AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) UpdateAgentInfrastructure(ctx context.Context, infrastructureID string, provider models.AIProvider) (*AIInfrastructureResult, error) {
	slog.Info("Updating AI infrastructure", "infrastructure_id", infrastructureID, "provider", provider)

	// Validate the configuration first
	if err := f.ValidateAgentConfig(provider); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	// For now, we'll implement update by destroying and recreating
	// TODO: In the future, this can be optimized to do in-place updates where possible

	// First, try to destroy existing infrastructure
	if err := f.DestroyAgentInfrastructure(ctx, infrastructureID); err != nil {
		slog.Warn("Failed to destroy existing infrastructure during update",
			"infrastructure_id", infrastructureID, "error", err)
		// Continue with creation anyway
	}

	// Create new infrastructure with updated configuration
	result, err := f.CreateAgentInfrastructure(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create updated infrastructure: %w", err)
	}

	slog.Info("AI infrastructure updated successfully",
		"old_infrastructure_id", infrastructureID,
		"new_infrastructure_id", result.AgentID,
		"provider", provider)

	return result, nil
}
