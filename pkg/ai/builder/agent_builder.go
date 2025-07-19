// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
)

// AgentBuilder defines how an agent is configured based on RAG metadata.
//
//go:generate mockgen -destination=./mocks/mock_agent_builder.go -mock_names=AgentBuilder=MockAgentBuilder -package=mocks . AgentBuilder
type AgentBuilder interface {
	// Build creates an agent connected to the RAG metadata (vector store).
	Build(ctx context.Context, ragID string) (string, string, error)

	// TearDown removes the agent and its associated resources.
	TearDown(ctx context.Context, agentID string, agentVersion string, ragID string) error
}

// AgentMetadata contains information needed to use or call the agent.
type AgentMetadata struct {
	// AgentID is the platform-specific identifier (e.g., Bedrock Agent ID).
	AgentID string

	// Endpoint is the URL or identifier you’ll hit to invoke the agent.
	Endpoint string

	// Provider specifies the backend (e.g., bedrock, openai, mock).
	Provider string
}
