// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
	"fmt"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
)

const (
	// ProviderName is the name of the provider used for building the RAG pipeline.
	ProviderName = "bedrock"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct {
	repoPath  string
	dataStore storage.DataStore
	rag       rag.RAG
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(
	repoPath string,
	storage storage.DataStore,
	rag rag.RAG,
) RAGBuilder {
	return &BedrockRAGBuilder{
		repoPath:  repoPath,
		dataStore: storage,
		rag:       rag,
	}
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(ctx context.Context) (string, error) {
	// TODO: Invoke Ensure RDS Postgres Schema Lambda

	// Create the RAG object
	kbID, err := b.rag.Create(ctx, b.getRDSTableName())
	if err != nil {
		return "", fmt.Errorf("failed to create RAG pipeline: %w", err)
	}

	// Upload the codebase to S3 if not already done
	err = b.dataStore.UploadDirectory(ctx, b.repoPath, b.repoPath)
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
	err = b.dataStore.DeleteDirectory(ctx, b.repoPath)
	if err != nil {
		return fmt.Errorf("failed to remove codebase from S3: %w", err)
	}

	// Delete the RAG pipeline
	err = b.rag.Delete(ctx, vectorStoreID)
	if err != nil {
		return fmt.Errorf("failed to delete RAG pipeline: %w", err)
	}

	return nil
}

// getRDSTableName returns the name of the RDS table used for vector storage.
func (b BedrockRAGBuilder) getRDSTableName() string {
	return b.repoPath
}
