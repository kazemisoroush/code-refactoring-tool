// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import "context"

// BedrockAgentBuilder is an implementation of AgentBuilder that uses AWS Bedrock for building agents.
type BedrockAgentBuilder struct {
}

// NewBedrockAgentBuilder creates a new instance of BedrockAgentBuilder.
func NewBedrockAgentBuilder() (AgentBuilder, error) {
	return BedrockAgentBuilder{}, nil
}

// Build implements AgentBuilder.
func (b BedrockAgentBuilder) Build(_ context.Context, _ *RAGMetadata) (AgentMetadata, error) {
	panic("unimplemented")
}
