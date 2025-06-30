package rag

import "context"

// RAG interface defines methods for creating and deleting RAG entries in a storage system.
//
//go:generate mockgen -destination=./mocks/mock_rag.go -mock_names=RAG=MockRAG -package=mocks . RAG
type RAG interface {
	Create(ctx context.Context, table string) (string, error)
	Delete(ctx context.Context, id string) error
}
