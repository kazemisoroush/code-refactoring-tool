// Package factory provides dynamic AI service creation based on agent configurations
package factory

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// TaskExecutionFactory creates AI resources on-demand for task execution
type TaskExecutionFactory interface {
	// CreateAgentForTask creates an agent for a specific task execution
	CreateAgentForTask(ctx context.Context, task *models.TaskWithFullContext) (string, string, error)

	// CreateRAGForCodebase creates a RAG pipeline for a specific codebase
	CreateRAGForCodebase(ctx context.Context, codebase *models.CodebaseWithContext) (string, error)

	// ValidateAgentConfig validates an agent's AI configuration
	ValidateAgentConfig(config models.AgentAIConfig) error

	// GetSupportedProviders returns list of supported AI providers
	GetSupportedProviders() []string
}

// DefaultTaskExecutionFactory implements the TaskExecutionFactory interface
type DefaultTaskExecutionFactory struct {
	// AWS configuration for Bedrock
	awsConfig aws.Config

	// Local AI configuration
	localConfig config.LocalAIConfig

	// Git configuration
	gitConfig config.GitConfig

	// Builder instances (created lazily)
	bedrockAgentBuilder builder.AgentBuilder
	bedrockRAGBuilder   builder.RAGBuilder
}

// NewTaskExecutionFactory creates a new dynamic factory for task execution
func NewTaskExecutionFactory(
	awsConfig aws.Config,
	localConfig config.LocalAIConfig,
	gitConfig config.GitConfig,
) TaskExecutionFactory {
	return &DefaultTaskExecutionFactory{
		awsConfig:   awsConfig,
		localConfig: localConfig,
		gitConfig:   gitConfig,
	}
}

// CreateAgentForTask creates an agent based on the task's agent configuration
func (f *DefaultTaskExecutionFactory) CreateAgentForTask(ctx context.Context, task *models.TaskWithFullContext) (string, string, error) {
	if task.Agent == nil {
		return "", "", fmt.Errorf("task agent not loaded in context")
	}

	agentConfig := task.Agent.AIConfig

	switch agentConfig.Provider {
	case models.AIProviderBedrock:
		return f.createBedrockAgent(ctx, agentConfig.Bedrock, task)
	case models.AIProviderLocal:
		return f.createLocalAgent(ctx, agentConfig.Local, task)
	default:
		return "", "", fmt.Errorf("unsupported AI provider: %s", agentConfig.Provider)
	}
}

// CreateRAGForCodebase creates a RAG pipeline based on codebase configuration
func (f *DefaultTaskExecutionFactory) CreateRAGForCodebase(ctx context.Context, cb *models.CodebaseWithContext) (string, error) {
	if cb.Project == nil {
		return "", fmt.Errorf("codebase project not loaded in context")
	}

	// Use project's default AI provider for RAG creation
	defaultProvider := cb.Project.Config.DefaultAIProvider
	if defaultProvider == "" {
		defaultProvider = string(models.AIProviderBedrock) // fallback
	}

	switch models.AIProvider(defaultProvider) {
	case models.AIProviderBedrock:
		return f.createBedrockRAG(ctx, cb)
	case models.AIProviderLocal:
		return f.createLocalRAG(ctx, cb)
	default:
		return "", fmt.Errorf("unsupported RAG provider: %s", defaultProvider)
	}
}

// ValidateAgentConfig validates an agent's AI configuration
func (f *DefaultTaskExecutionFactory) ValidateAgentConfig(config models.AgentAIConfig) error {
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
func (f *DefaultTaskExecutionFactory) GetSupportedProviders() []string {
	return []string{
		string(models.AIProviderBedrock),
		string(models.AIProviderLocal),
	}
}

// Private helper methods for creating provider-specific agents
func (f *DefaultTaskExecutionFactory) createBedrockAgent(ctx context.Context, config *models.BedrockAgentConfig, task *models.TaskWithFullContext) (string, string, error) {
	if config == nil {
		return "", "", fmt.Errorf("bedrock configuration is required")
	}

	// Initialize Bedrock agent builder if not already done
	if f.bedrockAgentBuilder == nil {
		f.bedrockAgentBuilder = builder.NewBedrockAgentBuilder(
			f.awsConfig,
			f.gitConfig.CodebaseURL,
			config.AgentServiceRoleARN,
		)
	}

	// First, create RAG for the task if codebase is specified
	var ragID string
	if task.Codebase != nil {
		ragID = *task.Agent.KnowledgeBaseID // Use existing knowledge base ID if available
		if ragID == "" {
			// Create RAG pipeline for the codebase
			var err error
			ragID, err = f.CreateRAGForCodebase(ctx, &models.CodebaseWithContext{
				Codebase: *task.Codebase,
				Project:  task.Project,
			})
			if err != nil {
				return "", "", fmt.Errorf("failed to create RAG for codebase: %w", err)
			}
		}
	}

	// Build agent with the RAG ID
	return f.bedrockAgentBuilder.Build(ctx, ragID)
}

func (f *DefaultTaskExecutionFactory) createLocalAgent(ctx context.Context, config *models.LocalAgentConfig, task *models.TaskWithFullContext) (string, string, error) {
	if config == nil {
		return "", "", fmt.Errorf("local configuration is required")
	}

	// Use existing LocalAgentBuilder
	builder := builder.NewLocalAgentBuilder(config.OllamaURL, config.Model)

	// We don't have a specific RAG ID for local, use a placeholder
	ragID := "local-rag-" + task.TaskID

	return builder.Build(ctx, ragID)
}

// Private helper methods for creating RAG instances
func (f *DefaultTaskExecutionFactory) createBedrockRAG(ctx context.Context, _ *models.CodebaseWithContext) (string, error) {
	// Initialize Bedrock RAG builder if not already done
	if f.bedrockRAGBuilder == nil {
		return "", fmt.Errorf("bedrock RAG builder not initialized - requires dataStore, storage, and rag dependencies")
	}

	// Build RAG pipeline
	return f.bedrockRAGBuilder.Build(ctx)
}

func (f *DefaultTaskExecutionFactory) createLocalRAG(ctx context.Context, _ *models.CodebaseWithContext) (string, error) {
	// Use existing LocalRAGBuilder
	// Default local configuration - in production this would come from config
	repoPath := "/tmp/local-repos" // Default path for local repos
	chromaURL := "http://localhost:8000"
	embeddingModel := "nomic-embed-text"

	builder := builder.NewLocalRAGBuilder(repoPath, chromaURL, embeddingModel)
	return builder.Build(ctx)
}

// Private validation methods
func (f *DefaultTaskExecutionFactory) validateBedrockConfig(config *models.BedrockAgentConfig) error {
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

func (f *DefaultTaskExecutionFactory) validateLocalConfig(config *models.LocalAgentConfig) error {
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
