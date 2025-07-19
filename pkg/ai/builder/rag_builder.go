// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
)

// RAGBuilder is responsible for building a RAG pipeline for a given code repository.
//
//go:generate mockgen -destination=./mocks/mock_rag_builder.go -mock_names=RAGBuilder=MockRAGBuilder -package=mocks . RAGBuilder
type RAGBuilder interface {
	// Build constructs the RAG pipeline from the provided repository.
	Build(ctx context.Context) (string, error)

	// TearDown cleans up any resources created during the RAG pipeline setup.
	TearDown(ctx context.Context, vectorStoreID string, ragID string) error
}
