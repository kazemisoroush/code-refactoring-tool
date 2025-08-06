package factory

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// DefaultAIInfrastructureFactory implements the AIInfrastructureFactory interface
type DefaultAIInfrastructureFactory struct {
	// AWS configuration for Bedrock
	awsConfig aws.Config

	// Base AI configuration from program config
	aiConfig config.AIConfig

	// Git configuration
	gitConfig config.GitConfig

	// Builder instances (created lazily)
	bedrockAgentBuilder builder.AgentBuilder
	bedrockRAGBuilder   builder.RAGBuilder
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
func (f *DefaultAIInfrastructureFactory) CreateAgentInfrastructure(ctx context.Context, provider models.AIProvider, repositoryURL string) (*AIInfrastructureResult, error) {
	switch provider {
	case models.AIProviderBedrock:
		return f.createBedrockInfrastructure(ctx, repositoryURL)
	case models.AIProviderLocal:
		return f.createLocalInfrastructure(ctx, repositoryURL)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
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

// GetSupportedProviders returns list of supported AI providers
func (f *DefaultAIInfrastructureFactory) GetSupportedProviders() []string {
	return []string{
		string(models.AIProviderBedrock),
		string(models.AIProviderLocal),
	}
}

// DestroyAgentInfrastructure cleans up AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) DestroyAgentInfrastructure(_ context.Context, infrastructureID string) error {
	// Implementation would depend on the provider and infrastructure type
	// For now, return a placeholder implementation
	return fmt.Errorf("infrastructure destruction not yet implemented for ID: %s", infrastructureID)
}

// Private helper methods for creating provider-specific infrastructure

func (f *DefaultAIInfrastructureFactory) createBedrockInfrastructure(ctx context.Context, repositoryURL string) (*AIInfrastructureResult, error) {
	config := f.aiConfig.Bedrock

	// Initialize Bedrock agent builder if not already done
	if f.bedrockAgentBuilder == nil {
		f.bedrockAgentBuilder = builder.NewBedrockAgentBuilder(
			f.awsConfig,
			repositoryURL,
			config.AgentServiceRoleARN,
		)
	}

	// Create RAG pipeline first
	var ragID string
	if f.bedrockRAGBuilder != nil {
		var err error
		ragID, err = f.bedrockRAGBuilder.Build(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create RAG infrastructure: %w", err)
		}
	}

	// Build agent with the RAG ID
	agentID, vectorStoreID, err := f.bedrockAgentBuilder.Build(ctx, ragID)
	if err != nil {
		return nil, fmt.Errorf("failed to create bedrock agent: %w", err)
	}

	return &AIInfrastructureResult{
		AgentID:         agentID,
		AgentVersion:    "v1.0.0", // This would come from Bedrock response
		KnowledgeBaseID: ragID,
		VectorStoreID:   vectorStoreID,
		Status:          models.AgentStatusInitializing,
		Metadata: map[string]interface{}{
			"provider":         "bedrock",
			"region":           config.Region,
			"foundation_model": config.FoundationModel,
		},
	}, nil
}

func (f *DefaultAIInfrastructureFactory) createLocalInfrastructure(ctx context.Context, repositoryURL string) (*AIInfrastructureResult, error) {
	config := f.aiConfig.Local

	// Use existing LocalAgentBuilder
	builder := builder.NewLocalAgentBuilder(config.OllamaURL, config.Model)

	// Create a RAG ID for local (simplified)
	ragID := "local-rag-" + repositoryURL

	agentID, vectorStoreID, err := builder.Build(ctx, ragID)
	if err != nil {
		return nil, fmt.Errorf("failed to create local agent: %w", err)
	}

	return &AIInfrastructureResult{
		AgentID:         agentID,
		AgentVersion:    "v1.0.0",
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

// UpdateAgentInfrastructure updates existing AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) UpdateAgentInfrastructure(ctx context.Context, infrastructureID string, provider models.AIProvider, repositoryURL string) (*AIInfrastructureResult, error) {
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
	result, err := f.CreateAgentInfrastructure(ctx, provider, repositoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create updated infrastructure: %w", err)
	}

	slog.Info("AI infrastructure updated successfully",
		"old_infrastructure_id", infrastructureID,
		"new_infrastructure_id", result.AgentID,
		"provider", provider)

	return result, nil
}
