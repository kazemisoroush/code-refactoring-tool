// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// RAGBuilder is responsible for building a RAG pipeline for a given code repository.
//
//go:generate mockgen -destination=./mocks/mock_ragbuilder.go -mock_names=RAGBuilder=MockRAGBuilder -package=mocks . RAGBuilder
type RAGBuilder interface {
	// Build constructs the RAG pipeline from the provided repository.
	Build(ctx context.Context, repo repository.Repository) (RAGMetadata, error)

	// TearDown cleans up any resources created during the RAG pipeline setup.
	TearDown(ctx context.Context, vectorStoreID string) error
}

// RAGMetadata represents the abstract result of a RAG pipeline setup.
type RAGMetadata struct {
	// VectorStoreID could be a Bedrock Knowledge Base ID, Pinecone Index, etc.
	VectorStoreID string

	// DataLocation is the URI (e.g., S3 bucket, GCS, local path) where the codebase was stored.
	DataLocation string

	// Provider describes which system was used (e.g., "bedrock", "openai", "mock")
	Provider string
}
