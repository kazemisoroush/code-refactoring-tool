// Package repository provides data access implementations
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	conf "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// PostgresCodebaseRepository implements CodebaseRepository using PostgreSQL
type PostgresCodebaseRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresCodebaseRepository creates a new PostgreSQL codebase repository
func NewPostgresCodebaseRepository(config PostgresConfig, tableName string) (CodebaseRepository, error) {
	if tableName == "" {
		tableName = conf.DefaultCodebasesTableName
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	repo := &PostgresCodebaseRepository{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := repo.createTableIfNotExists(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return repo, nil
}

// createTableIfNotExists creates the codebases table if it doesn't exist
func (r *PostgresCodebaseRepository) createTableIfNotExists() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			codebase_id UUID PRIMARY KEY,
			project_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			provider VARCHAR(50) NOT NULL,
			url TEXT NOT NULL,
			default_branch VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			metadata JSONB,
			tags JSONB,
			CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_%s_project_id ON %s(project_id);
		CREATE INDEX IF NOT EXISTS idx_%s_provider ON %s(provider);
		CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at);
		CREATE INDEX IF NOT EXISTS idx_%s_tags ON %s USING GIN(tags);
	`, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName)

	_, err := r.db.Exec(query)
	return err
}

// CreateCodebase creates a new codebase record
func (r *PostgresCodebaseRepository) CreateCodebase(ctx context.Context, codebase *models.Codebase) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (codebase_id, project_id, name, provider, url, default_branch, created_at, updated_at, metadata, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName)

	metadataJSON, err := json.Marshal(codebase.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	tagsJSON, err := json.Marshal(codebase.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		codebase.CodebaseID,
		codebase.ProjectID,
		codebase.Name,
		codebase.Provider,
		codebase.URL,
		codebase.DefaultBranch,
		codebase.CreatedAt,
		codebase.UpdatedAt,
		metadataJSON,
		tagsJSON,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("codebase with ID %s already exists", codebase.CodebaseID)
			case "23503": // foreign_key_violation
				return fmt.Errorf("project with ID %s does not exist", codebase.ProjectID)
			}
		}
		return fmt.Errorf("failed to create codebase: %w", err)
	}

	return nil
}

// GetCodebase retrieves a codebase by ID
func (r *PostgresCodebaseRepository) GetCodebase(ctx context.Context, codebaseID string) (*models.Codebase, error) {
	query := fmt.Sprintf(`
		SELECT codebase_id, project_id, name, provider, url, default_branch, created_at, updated_at, metadata, tags
		FROM %s
		WHERE codebase_id = $1
	`, r.tableName)

	var codebase models.Codebase
	var metadataJSON, tagsJSON []byte

	err := r.db.QueryRowContext(ctx, query, codebaseID).Scan(
		&codebase.CodebaseID,
		&codebase.ProjectID,
		&codebase.Name,
		&codebase.Provider,
		&codebase.URL,
		&codebase.DefaultBranch,
		&codebase.CreatedAt,
		&codebase.UpdatedAt,
		&metadataJSON,
		&tagsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("codebase not found")
		}
		return nil, fmt.Errorf("failed to get codebase: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &codebase.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &codebase.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return &codebase, nil
}

// UpdateCodebase updates an existing codebase
func (r *PostgresCodebaseRepository) UpdateCodebase(ctx context.Context, codebase *models.Codebase) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = $2, default_branch = $3, updated_at = $4, metadata = $5, tags = $6
		WHERE codebase_id = $1
	`, r.tableName)

	metadataJSON, err := json.Marshal(codebase.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	tagsJSON, err := json.Marshal(codebase.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		codebase.CodebaseID,
		codebase.Name,
		codebase.DefaultBranch,
		codebase.UpdatedAt,
		metadataJSON,
		tagsJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update codebase: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("codebase not found")
	}

	return nil
}

// DeleteCodebase deletes a codebase by ID
func (r *PostgresCodebaseRepository) DeleteCodebase(ctx context.Context, codebaseID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE codebase_id = $1`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, codebaseID)
	if err != nil {
		return fmt.Errorf("failed to delete codebase: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("codebase not found")
	}

	return nil
}

// ListCodebases lists codebases with optional filtering and pagination
func (r *PostgresCodebaseRepository) ListCodebases(ctx context.Context, filter CodebaseFilter) ([]*models.Codebase, string, error) {
	// Build the query dynamically based on filters
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := fmt.Sprintf(`
		SELECT codebase_id, project_id, name, provider, url, default_branch, created_at, updated_at, metadata, tags
		FROM %s
	`, r.tableName)

	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *filter.ProjectID)
		argIndex++
	}

	if filter.Provider != nil {
		conditions = append(conditions, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, *filter.Provider)
		argIndex++
	}

	if filter.TagFilter != nil {
		// Parse tag filter (format: key:value)
		parts := strings.SplitN(*filter.TagFilter, ":", 2)
		if len(parts) == 2 {
			conditions = append(conditions, fmt.Sprintf("tags->>$%d = $%d", argIndex, argIndex+1))
			args = append(args, parts[0], parts[1])
			argIndex += 2
		}
	}

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering and pagination
	baseQuery += " ORDER BY created_at DESC"

	maxResults := 50 // default
	if filter.MaxResults != nil {
		maxResults = *filter.MaxResults
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, maxResults+1) // Get one extra to determine if there's a next page

	if filter.NextToken != nil {
		// Parse next token (could be a timestamp or cursor)
		baseQuery += fmt.Sprintf(" OFFSET $%d", argIndex+1)
		args = append(args, *filter.NextToken)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list codebases: %w", err)
	}
	defer func() {
		_ = rows.Close() // Ignore close error as we're already handling the main error
	}()

	var codebases []*models.Codebase

	for rows.Next() {
		var codebase models.Codebase
		var metadataJSON, tagsJSON []byte

		err := rows.Scan(
			&codebase.CodebaseID,
			&codebase.ProjectID,
			&codebase.Name,
			&codebase.Provider,
			&codebase.URL,
			&codebase.DefaultBranch,
			&codebase.CreatedAt,
			&codebase.UpdatedAt,
			&metadataJSON,
			&tagsJSON,
		)

		if err != nil {
			return nil, "", fmt.Errorf("failed to scan codebase: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &codebase.Metadata); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &codebase.Tags); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		codebases = append(codebases, &codebase)
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error iterating rows: %w", err)
	}

	// Determine next token
	var nextToken string
	if len(codebases) > maxResults {
		// Remove the extra item
		codebases = codebases[:maxResults]
		// Set next token (simple offset-based pagination)
		currentOffset := 0
		if filter.NextToken != nil {
			// Parse current offset from token
			if _, err := fmt.Sscanf(*filter.NextToken, "%d", &currentOffset); err != nil {
				// If parsing fails, use 0 as default
				currentOffset = 0
			}
		}
		nextToken = fmt.Sprintf("%d", currentOffset+maxResults)
	}

	return codebases, nextToken, nil
}

// CodebaseExists checks if a codebase exists
func (r *PostgresCodebaseRepository) CodebaseExists(ctx context.Context, codebaseID string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE codebase_id = $1)`, r.tableName)

	var exists bool
	err := r.db.QueryRowContext(ctx, query, codebaseID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check codebase existence: %w", err)
	}

	return exists, nil
}

// GetCodebasesByProject gets all codebases for a specific project
func (r *PostgresCodebaseRepository) GetCodebasesByProject(ctx context.Context, projectID string) ([]*models.Codebase, error) {
	query := fmt.Sprintf(`
		SELECT codebase_id, project_id, name, provider, url, default_branch, created_at, updated_at, metadata, tags
		FROM %s
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, r.tableName)

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebases by project: %w", err)
	}
	defer func() {
		_ = rows.Close() // Ignore close error as we're already handling the main error
	}()

	var codebases []*models.Codebase

	for rows.Next() {
		var codebase models.Codebase
		var metadataJSON, tagsJSON []byte

		err := rows.Scan(
			&codebase.CodebaseID,
			&codebase.ProjectID,
			&codebase.Name,
			&codebase.Provider,
			&codebase.URL,
			&codebase.DefaultBranch,
			&codebase.CreatedAt,
			&codebase.UpdatedAt,
			&metadataJSON,
			&tagsJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan codebase: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &codebase.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &codebase.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		codebases = append(codebases, &codebase)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return codebases, nil
}
