// Package storage provides interfaces and implementations for persisting data
// and managing storage operations like schema creation in backing services
// such as RDS or S3.
package storage

import "context"

// Storage defines an interface for storage-related operations.
//
// It currently includes a method for ensuring that a schema (e.g., table)
// exists in the backing data store.
//
//go:generate mockgen -destination=./mocks/mock_storage.go -mock_names=Storage=MockStorage -package=mocks . Storage
type Storage interface {
	// EnsureSchema ensures that the schema (e.g., a table) exists in the storage layer.
	// This method should be idempotent — i.e., safe to call multiple times without side effects.
	//
	// Parameters:
	//   - ctx: Context for managing cancellation and timeouts.
	//   - databaseName: Name of the database to create or verify.
	//   - tableName: Name of the table or schema to create or verify.
	//
	// Returns:
	//   - error: Any error encountered while ensuring the schema.
	EnsureSchema(ctx context.Context, databaseName string, tableName string) error
}
