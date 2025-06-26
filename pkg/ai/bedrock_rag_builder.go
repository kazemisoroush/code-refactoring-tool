// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct{}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder() (RAGBuilder, error) {
	return BedrockRAGBuilder{}, nil
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(_ context.Context, _ repository.Repository) (RAGMetadata, error) {
	panic("unimplemented")
}
