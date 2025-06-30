// Package vector provides an interface for vector data stores, specifically for ensuring and dropping schemas in a vector store.
package vector

import "context"

// Storage defines the interface for a vector store that can ensure its schema and drop it.
//
//go:generate mockgen -destination=./mocks/mock_storage.go -mock_names=Storage=MockStorage -package=mocks . Storage
type Storage interface {
	EnsureSchema(ctx context.Context, tableName string) error
	DropSchema(ctx context.Context, tableName string) error
}
