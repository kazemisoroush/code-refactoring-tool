// Package storage provides an implementation of a vector data store using AWS RDS Postgres.
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// RDSVector is an interface for vector data stores.
type RDSVector struct {
	conn *pgx.Conn
}

// NewRDSVector creates a new instance of RDSVectorStore with the provided AWS configuration and parameters.
func NewRDSVector(ctx context.Context, cfg config.RDSPostgres) (Vector, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DatabaseName,
	)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	return &RDSVector{conn: conn}, nil
}

// EnsureSchema creates the table if it doesn't exist.
func (r *RDSVector) EnsureSchema(ctx context.Context, tableName string) error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			text TEXT,
			embedding VECTOR,
			metadata JSON
		)`, tableName)

	_, err := r.conn.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create/check table: %w", err)
	}

	return nil
}

// DropSchema drops the table if it exists.
func (r *RDSVector) DropSchema(ctx context.Context, tableName string) error {
	dropTableSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
	_, err := r.conn.Exec(ctx, dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	return nil
}
