// Package storage provides an interface for vector data stores, specifically for ensuring and dropping schemas in a vector store.
package storage

import "context"

// Vector defines the interface for a vector store that can ensure its schema and drop it.
//
//go:generate mockgen -destination=./mocks/mock_vector.go -mock_names=Vector=MockVector -package=mocks . Vector
type Vector interface {
	EnsureSchema(ctx context.Context, tableName string) error
	DropSchema(ctx context.Context, tableName string) error
}
