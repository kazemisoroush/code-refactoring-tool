// Package factory provides interfaces and implementations for creating and managing AI infrastructure.
package factory

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// AIInfrastructureFactory creates AI infrastructure on-demand based on configuration
//
//go:generate mockgen -destination=./mocks/mock_ai_infrastructure_factory.go -mock_names=AIInfrastructureFactory=MockAIInfrastructureFactory -package=mocks . AIInfrastructureFactory
type AIInfrastructureFactory interface {
	// CreateAgentInfrastructure creates AI infrastructure for an agent
	CreateAgentInfrastructure(ctx context.Context, provider models.AIProvider, repositoryURL string) (*AIInfrastructureResult, error)

	// UpdateAgentInfrastructure updates existing AI infrastructure for an agent
	UpdateAgentInfrastructure(ctx context.Context, infrastructureID string, provider models.AIProvider, repositoryURL string) (*AIInfrastructureResult, error)

	// ValidateAgentConfig validates an agent's AI provider configuration
	ValidateAgentConfig(provider models.AIProvider) error

	// GetSupportedProviders returns list of supported AI providers
	GetSupportedProviders() []string

	// DestroyAgentInfrastructure cleans up AI infrastructure for an agent
	DestroyAgentInfrastructure(ctx context.Context, infrastructureID string) error
}

// AIInfrastructureResult contains the result of creating AI infrastructure
type AIInfrastructureResult struct {
	// AgentID is the unique identifier for the created agent
	AgentID string

	// AgentVersion is the version of the created agent
	AgentVersion string

	// KnowledgeBaseID is the ID of the knowledge base (if created)
	KnowledgeBaseID string

	// VectorStoreID is the ID of the vector store (if created)
	VectorStoreID string

	// Status indicates the current status of the infrastructure
	Status models.AgentStatus

	// Metadata contains provider-specific metadata
	Metadata map[string]interface{}
}

// AgentCreationRequest contains parameters for creating an agent
type AgentCreationRequest struct {
	AgentName     string
	RepositoryURL string
	Branch        string
	AIProvider    models.AIProvider
}
