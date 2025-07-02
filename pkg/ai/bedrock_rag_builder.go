// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"
	"fmt"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

const (
	// ProviderName is the name of the provider used for building the RAG pipeline.
	ProviderName = "bedrock"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct {
	repository    repository.Repository
	dataStore     storage.DataStore
	rag           rag.RAG
	vectorStorage storage.Vector
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(
	repository repository.Repository,
	storage storage.DataStore,
	vectorDataStore storage.Vector,
	rag rag.RAG,
) RAGBuilder {
	return &BedrockRAGBuilder{
		vectorStorage: vectorDataStore,
		repository:    repository,
		dataStore:     storage,
		rag:           rag,
	}
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(ctx context.Context) (string, error) {
	// Create RDS Aurora table if it doesn't exist
	err := b.vectorStorage.EnsureSchema(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return "", fmt.Errorf("failed to ensure RDS Aurora table schema: %w", err)
	}

	// Create the RAG object
	kbID, err := b.rag.Create(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return "", fmt.Errorf("failed to create RAG pipeline: %w", err)
	}

	// Upload the codebase to S3 if not already done
	repoPath := b.repository.GetPath()
	err = b.dataStore.UploadDirectory(ctx, repoPath, repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to upload codebase to S3: %w", err)
	}

	// Create data source for the codebase in the RAG pipeline
	_, err = b.dataStore.Create(ctx, kbID)
	if err != nil {
		return "", fmt.Errorf("failed to create data source: %w", err)
	}

	return kbID, nil
}

// TearDown implements RAGBuilder.
func (b BedrockRAGBuilder) TearDown(ctx context.Context, vectorStoreID string, ragID string) error {
	// Delete the data source from the RAG pipeline if it exists
	err := b.dataStore.Detele(ctx, vectorStoreID, ragID)
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	// Remove the codebase from S3 if needed
	repoPath := b.repository.GetPath()
	err = b.dataStore.DeleteDirectory(ctx, repoPath)
	if err != nil {
		return fmt.Errorf("failed to remove codebase from S3: %w", err)
	}

	// Delete the RAG pipeline
	err = b.rag.Delete(ctx, vectorStoreID)
	if err != nil {
		return fmt.Errorf("failed to delete RAG pipeline: %w", err)
	}

	// Drop the RDS Aurora table used for vector storage
	err = b.vectorStorage.DropSchema(ctx, b.getRDSAuroraTableName())
	if err != nil {
		return fmt.Errorf("failed to drop RDS Aurora table: %w", err)
	}

	return nil
}

// getRDSAuroraTableName returns the name of the RDS table used for vector storage.
func (b BedrockRAGBuilder) getRDSAuroraTableName() string {
	return b.repository.GetPath()
}
