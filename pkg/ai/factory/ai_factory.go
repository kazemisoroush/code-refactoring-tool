// Package factory provides factory functions for creating AI implementations based on configuration.
package factory

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// AIProviderFactory interface defines methods for creating AI implementations based on provider configuration.
//
//go:generate mockgen -destination=./mocks/mock_ai_provider_factory.go -mock_names=AIProviderFactory=MockAIProviderFactory -package=mocks . AIProviderFactory
type AIProviderFactory interface {
	// CreateRAGBuilder creates a RAGBuilder implementation based on the AI configuration
	CreateRAGBuilder(aiConfig *config.AIConfiguration, repoPath string) (builder.RAGBuilder, error)

	// CreateAgentBuilder creates an AgentBuilder implementation based on the AI configuration
	CreateAgentBuilder(aiConfig *config.AIConfiguration, repoPath string) (builder.AgentBuilder, error)

	// ValidateAIConfiguration validates the AI configuration and returns appropriate errors
	ValidateAIConfiguration(aiConfig *config.AIConfiguration) error
}

// DefaultAIProviderFactory is the default implementation of AIProviderFactory
// DefaultAIProviderFactory is the default implementation of AIProviderFactory
type DefaultAIProviderFactory struct {
	awsConfig                   aws.Config
	s3BucketName                string
	knowledgeBaseServiceRoleARN string
	agentServiceRoleARN         string
	localAIEnabled              bool
	localOllamaURL              string
	localModel                  string
	localChromaURL              string
	localEmbeddingModel         string
	dataStore                   storage.DataStore
	storage                     storage.Storage
	ragService                  rag.RAG
}

// NewAIProviderFactory creates a new AI provider factory
func NewAIProviderFactory(cfg *config.Config, dataStore storage.DataStore, storageService storage.Storage, ragService rag.RAG) AIProviderFactory {
	return &DefaultAIProviderFactory{
		awsConfig:                   cfg.AWSConfig,
		s3BucketName:                cfg.S3BucketName,
		knowledgeBaseServiceRoleARN: cfg.KnowledgeBaseServiceRoleARN,
		agentServiceRoleARN:         cfg.AgentServiceRoleARN,
		localAIEnabled:              cfg.LocalAI.Enabled,
		localOllamaURL:              cfg.LocalAI.OllamaURL,
		localModel:                  cfg.LocalAI.Model,
		localChromaURL:              cfg.LocalAI.ChromaURL,
		localEmbeddingModel:         cfg.LocalAI.EmbeddingModel,
		dataStore:                   dataStore,
		storage:                     storageService,
		ragService:                  ragService,
	}
}

// CreateRAGBuilder creates a RAGBuilder implementation based on the AI configuration
func (f *DefaultAIProviderFactory) CreateRAGBuilder(aiConfig *config.AIConfiguration, repoPath string) (builder.RAGBuilder, error) {
	provider := f.getEffectiveProvider(aiConfig)

	switch provider {
	case config.AIProviderLocal:
		localConfig := f.getLocalConfig(aiConfig)
		return builder.NewLocalRAGBuilder(repoPath, localConfig.ChromaURL, localConfig.EmbeddingModel), nil
	case config.AIProviderBedrock:
		// Use existing Bedrock RAG builder creation logic
		return builder.NewBedrockRAGBuilder(repoPath, f.dataStore, f.storage, f.ragService), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// CreateAgentBuilder creates an AgentBuilder implementation based on the AI configuration
func (f *DefaultAIProviderFactory) CreateAgentBuilder(aiConfig *config.AIConfiguration, repoPath string) (builder.AgentBuilder, error) {
	provider := f.getEffectiveProvider(aiConfig)

	switch provider {
	case config.AIProviderLocal:
		localConfig := f.getLocalConfig(aiConfig)
		return builder.NewLocalAgentBuilder(localConfig.OllamaURL, localConfig.Model), nil
	case config.AIProviderBedrock:
		// Use existing Bedrock agent builder creation logic
		bedrockConfig := f.getBedrockConfig(aiConfig)
		return builder.NewBedrockAgentBuilder(f.awsConfig, repoPath, bedrockConfig.AgentServiceRoleARN), nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// getEffectiveProvider determines which AI provider to use based on configuration hierarchy
func (f *DefaultAIProviderFactory) getEffectiveProvider(aiConfig *config.AIConfiguration) string {
	// 1. User-provided configuration takes precedence
	if aiConfig != nil && aiConfig.GetProvider() != "" {
		return aiConfig.GetProvider()
	}

	// 2. Platform-level configuration from environment
	if f.localAIEnabled {
		return config.AIProviderLocal
	}

	// 3. Default to Bedrock for SaaS
	return config.AIProviderBedrock
}

// getLocalConfig extracts local AI configuration with fallbacks
func (f *DefaultAIProviderFactory) getLocalConfig(aiConfig *config.AIConfiguration) *config.LocalAIRequestConfig {
	// Start with platform defaults
	localConfig := &config.LocalAIRequestConfig{
		OllamaURL:      f.localOllamaURL,
		Model:          f.localModel,
		ChromaURL:      f.localChromaURL,
		EmbeddingModel: f.localEmbeddingModel,
	}

	// Override with user-provided configuration if available
	if aiConfig != nil && aiConfig.Local != nil {
		if aiConfig.Local.OllamaURL != "" {
			localConfig.OllamaURL = aiConfig.Local.OllamaURL
		}
		if aiConfig.Local.Model != "" {
			localConfig.Model = aiConfig.Local.Model
		}
		if aiConfig.Local.ChromaURL != "" {
			localConfig.ChromaURL = aiConfig.Local.ChromaURL
		}
		if aiConfig.Local.EmbeddingModel != "" {
			localConfig.EmbeddingModel = aiConfig.Local.EmbeddingModel
		}
	}

	return localConfig
}

// getBedrockConfig extracts Bedrock AI configuration with fallbacks
func (f *DefaultAIProviderFactory) getBedrockConfig(aiConfig *config.AIConfiguration) *config.BedrockAIRequestConfig {
	// Start with platform defaults
	bedrockConfig := &config.BedrockAIRequestConfig{
		KnowledgeBaseServiceRoleARN: f.knowledgeBaseServiceRoleARN,
		AgentServiceRoleARN:         f.agentServiceRoleARN,
		Region:                      f.awsConfig.Region,
	}

	// Override with user-provided configuration if available
	if aiConfig != nil && aiConfig.Bedrock != nil {
		if aiConfig.Bedrock.KnowledgeBaseServiceRoleARN != "" {
			bedrockConfig.KnowledgeBaseServiceRoleARN = aiConfig.Bedrock.KnowledgeBaseServiceRoleARN
		}
		if aiConfig.Bedrock.AgentServiceRoleARN != "" {
			bedrockConfig.AgentServiceRoleARN = aiConfig.Bedrock.AgentServiceRoleARN
		}
		if aiConfig.Bedrock.Region != "" {
			bedrockConfig.Region = aiConfig.Bedrock.Region
		}
		if aiConfig.Bedrock.FoundationModel != "" {
			bedrockConfig.FoundationModel = aiConfig.Bedrock.FoundationModel
		}
	}

	return bedrockConfig
}

// ValidateAIConfiguration validates the AI configuration and returns appropriate errors
func (f *DefaultAIProviderFactory) ValidateAIConfiguration(aiConfig *config.AIConfiguration) error {
	if aiConfig == nil {
		return nil // Use platform defaults
	}

	// Validate provider
	if err := aiConfig.ValidateProvider(); err != nil {
		return err
	}

	// Provider-specific validation
	switch aiConfig.GetProvider() {
	case config.AIProviderLocal:
		return f.validateLocalConfig(aiConfig.Local)
	case config.AIProviderBedrock:
		return f.validateBedrockConfig(aiConfig.Bedrock)
	case config.AIProviderOpenAI:
		return fmt.Errorf("OpenAI provider not yet implemented")
	}

	return nil
}

// validateLocalConfig validates local AI configuration
func (f *DefaultAIProviderFactory) validateLocalConfig(localConfig *config.LocalAIRequestConfig) error {
	if localConfig == nil {
		return nil // Use platform defaults
	}

	// Basic URL validation could be added here
	// For now, we'll rely on runtime connectivity checks
	return nil
}

// validateBedrockConfig validates Bedrock AI configuration
func (f *DefaultAIProviderFactory) validateBedrockConfig(bedrockConfig *config.BedrockAIRequestConfig) error {
	if bedrockConfig == nil {
		return nil // Use platform defaults
	}

	// ARN format validation could be added here
	// For now, we'll rely on AWS SDK validation
	return nil
}
