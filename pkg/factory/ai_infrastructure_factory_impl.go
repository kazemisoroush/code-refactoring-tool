package factory

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// DefaultAIInfrastructureFactory implements the AIInfrastructureFactory interface
type DefaultAIInfrastructureFactory struct {
	// AWS configuration for Bedrock
	awsConfig aws.Config

	// Git configuration
	gitConfig config.GitConfig

	// Builder instances (created lazily)
	bedrockAgentBuilder builder.AgentBuilder
	bedrockRAGBuilder   builder.RAGBuilder
}

// NewAIInfrastructureFactory creates a new AI infrastructure factory
func NewAIInfrastructureFactory(awsConfig aws.Config, gitConfig config.GitConfig) AIInfrastructureFactory {
	return &DefaultAIInfrastructureFactory{
		awsConfig: awsConfig,
		gitConfig: gitConfig,
	}
}

// CreateAgentInfrastructure creates AI infrastructure for an agent
func (f *DefaultAIInfrastructureFactory) CreateAgentInfrastructure(ctx context.Context, config *models.AgentAIConfig, repositoryURL string) (*AIInfrastructureResult, error) {
	if config == nil {
		return nil, fmt.Errorf("agent AI configuration is required")
	}

	switch config.Provider {
	case models.AIProviderBedrock:
		return f.createBedrockInfrastructure(ctx, config.Bedrock, repositoryURL)
	case models.AIProviderLocal:
		return f.createLocalInfrastructure(ctx, config.Local, repositoryURL)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", config.Provider)
	}
}

// ValidateAgentConfig validates an agent's AI configuration
func (f *DefaultAIInfrastructureFactory) ValidateAgentConfig(config *models.AgentAIConfig) error {
	if config == nil {
		return fmt.Errorf("agent AI configuration is required")
	}

	switch config.Provider {
	case models.AIProviderBedrock:
		return f.validateBedrockConfig(config.Bedrock)
	case models.AIProviderLocal:
		return f.validateLocalConfig(config.Local)
	default:
		return fmt.Errorf("unsupported AI provider: %s", config.Provider)
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

func (f *DefaultAIInfrastructureFactory) createBedrockInfrastructure(ctx context.Context, config *models.BedrockAgentConfig, repositoryURL string) (*AIInfrastructureResult, error) {
	if config == nil {
		return nil, fmt.Errorf("bedrock configuration is required")
	}

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

func (f *DefaultAIInfrastructureFactory) createLocalInfrastructure(ctx context.Context, config *models.LocalAgentConfig, repositoryURL string) (*AIInfrastructureResult, error) {
	if config == nil {
		return nil, fmt.Errorf("local configuration is required")
	}

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
func (f *DefaultAIInfrastructureFactory) validateBedrockConfig(config *models.BedrockAgentConfig) error {
	if config == nil {
		return fmt.Errorf("bedrock configuration is required")
	}
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

func (f *DefaultAIInfrastructureFactory) validateLocalConfig(config *models.LocalAgentConfig) error {
	if config == nil {
		return fmt.Errorf("local configuration is required")
	}
	if config.OllamaURL == "" {
		return fmt.Errorf("ollama URL is required")
	}
	if config.Model == "" {
		return fmt.Errorf("local model is required")
	}
	return nil
}
