// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"
	"fmt"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/vector"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

const (
	// ProviderName is the name of the provider used for building the RAG pipeline.
	ProviderName = "bedrock"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct {
	repository      repository.Repository
	storage         storage.Storage
	rag             rag.RAG
	vectorDataStore vector.Storage
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(
	repository repository.Repository,
	storage storage.Storage,
	vectorDataStore vector.Storage,
	rag rag.RAG,
) RAGBuilder {
	return &BedrockRAGBuilder{
		vectorDataStore: vectorDataStore,
		repository:      repository,
		storage:         storage,
		rag:             rag,
	}
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(ctx context.Context) (string, error) {
	// Upload the codebase to S3 if not already done
	repoPath := b.repository.GetPath()
	err := b.storage.UploadDirectory(ctx, repoPath, repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to upload codebase to S3: %w", err)
	}

	// Create RDS Aurora table if it doesn't exist
	err = b.vectorDataStore.EnsureSchema(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return "", fmt.Errorf("failed to ensure RDS Aurora table schema: %w", err)
	}

	// Create the RAG object
	kbID, err := b.rag.Create(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return "", fmt.Errorf("failed to create RAG pipeline: %w", err)
	}

	return kbID, nil
}

// TearDown implements RAGBuilder.
func (b BedrockRAGBuilder) TearDown(ctx context.Context, vectorStoreID string) error {
	// Remove the codebase from S3 if needed
	repoPath := b.repository.GetPath()
	err := b.storage.DeleteDirectory(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to remove codebase from S3: %w", err)
	}

	// Delete the RAG pipeline
	err = b.rag.Delete(ctx, vectorStoreID)
	if err != nil {
		return fmt.Errorf("failed to delete RAG pipeline: %w", err)
	}

	// Drop the RDS Aurora table used for vector storage
	err = b.vectorDataStore.DropSchema(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return fmt.Errorf("failed to drop RDS Aurora table: %w", err)
	}

	return nil
}

// getRDSAuroraTableName returns the name of the RDS table used for vector storage.
func (b BedrockRAGBuilder) getRDSAuroraTableName() string {
	return b.repository.GetPath()
}
