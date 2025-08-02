// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// LocalAgentBuilder implements the AgentBuilder interface for local development.
type LocalAgentBuilder struct {
	ollamaURL string
	model     string
}

// NewLocalAgentBuilder creates a new instance of LocalAgentBuilder.
func NewLocalAgentBuilder(ollamaURL, model string) AgentBuilder {
	return &LocalAgentBuilder{
		ollamaURL: ollamaURL,
		model:     model,
	}
}

// Build implements the AgentBuilder interface by creating a local agent configuration.
// For local development, we generate a unique agent ID and return it along with version info.
func (l *LocalAgentBuilder) Build(_ context.Context, ragID string) (string, string, error) {
	agentID := uuid.New().String()
	agentVersion := "v1.0.0"

	// In a real implementation, you might:
	// 1. Configure the agent with the RAG context
	// 2. Set up any necessary local resources
	// 3. Initialize the connection to the vector store

	// For now, we simulate successful agent creation
	fmt.Printf("Local agent created with ID: %s, version: %s, using RAG: %s\n", agentID, agentVersion, ragID)

	return agentID, agentVersion, nil
}

// TearDown implements the AgentBuilder interface by cleaning up local agent resources.
func (l *LocalAgentBuilder) TearDown(_ context.Context, agentID string, agentVersion string, ragID string) error {
	// In a real implementation, you might:
	// 1. Clean up any local resources
	// 2. Disconnect from the vector store
	// 3. Remove temporary files or configurations

	fmt.Printf("Local agent torn down - ID: %s, version: %s, RAG: %s\n", agentID, agentVersion, ragID)

	return nil
}
