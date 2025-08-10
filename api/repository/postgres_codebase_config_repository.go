// Package repository provides PostgreSQL implementation for codebase configuration data operations
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	conf "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/lib/pq"
)

// PostgresCodebaseConfigRepository implements CodebaseConfigRepository using PostgreSQL
type PostgresCodebaseConfigRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresCodebaseConfigRepository creates a new PostgreSQL codebase configuration repository
func NewPostgresCodebaseConfigRepository(config PostgresConfig, tableName string) (CodebaseConfigRepository, error) {
	if tableName == "" {
		tableName = conf.DefaultCodebaseConfigsTableName
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
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	repo := &PostgresCodebaseConfigRepository{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := repo.CreateTable(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return repo, nil
}

// NewPostgresCodebaseConfigRepositoryWithDB creates a new PostgreSQL codebase configuration repository with an existing DB connection
// This is primarily used for testing with mock databases
func NewPostgresCodebaseConfigRepositoryWithDB(db *sql.DB, tableName string) CodebaseConfigRepository {
	if tableName == "" {
		tableName = "codebase_configs"
	}

	return &PostgresCodebaseConfigRepository{
		db:        db,
		tableName: tableName,
	}
}

// Close closes the database connection
func (r *PostgresCodebaseConfigRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateCodebaseConfig creates a new codebase configuration record in PostgreSQL
func (r *PostgresCodebaseConfigRepository) CreateCodebaseConfig(ctx context.Context, config *CodebaseConfigRecord) error {
	// Convert maps and structs to JSON for storage
	tagsJSON, err := json.Marshal(config.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			config_id, name, description, provider, url, default_branch,
			status, created_at, updated_at, tags, metadata, config
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, r.tableName)

	_, err = r.db.ExecContext(ctx, query,
		config.ConfigID,
		config.Name,
		config.Description,
		config.Provider,
		config.URL,
		config.DefaultBranch,
		config.Status,
		config.CreatedAt,
		config.UpdatedAt,
		tagsJSON,
		metadataJSON,
		configJSON,
	)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("codebase configuration with ID %s already exists", config.ConfigID)
		}
		return fmt.Errorf("failed to create codebase configuration in PostgreSQL: %w", err)
	}

	return nil
}

// GetCodebaseConfig retrieves a codebase configuration by ID from PostgreSQL
func (r *PostgresCodebaseConfigRepository) GetCodebaseConfig(ctx context.Context, configID string) (*CodebaseConfigRecord, error) {
	query := fmt.Sprintf(`
		SELECT config_id, name, description, provider, url, default_branch,
			   status, created_at, updated_at, tags, metadata, config
		FROM %s WHERE config_id = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, configID)

	var config CodebaseConfigRecord
	var description sql.NullString
	var tagsJSON, metadataJSON, configJSON []byte

	err := row.Scan(
		&config.ConfigID,
		&config.Name,
		&description,
		&config.Provider,
		&config.URL,
		&config.DefaultBranch,
		&config.Status,
		&config.CreatedAt,
		&config.UpdatedAt,
		&tagsJSON,
		&metadataJSON,
		&configJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("codebase configuration not found")
		}
		return nil, fmt.Errorf("failed to get codebase configuration from PostgreSQL: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		config.Description = &description.String
	}

	// Unmarshal JSON fields
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &config.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &config.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	return &config, nil
}

// UpdateCodebaseConfig updates an existing codebase configuration record in PostgreSQL
func (r *PostgresCodebaseConfigRepository) UpdateCodebaseConfig(ctx context.Context, config *CodebaseConfigRecord) error {
	// Convert maps and structs to JSON for storage
	tagsJSON, err := json.Marshal(config.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(config.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s SET 
			name = $2, description = $3, provider = $4, url = $5, default_branch = $6,
			status = $7, updated_at = $8, tags = $9, metadata = $10, config = $11
		WHERE config_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query,
		config.ConfigID,
		config.Name,
		config.Description,
		config.Provider,
		config.URL,
		config.DefaultBranch,
		config.Status,
		config.UpdatedAt,
		tagsJSON,
		metadataJSON,
		configJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update codebase configuration in PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("codebase configuration with ID %s does not exist", config.ConfigID)
	}

	return nil
}

// DeleteCodebaseConfig deletes a codebase configuration by ID from PostgreSQL
func (r *PostgresCodebaseConfigRepository) DeleteCodebaseConfig(ctx context.Context, configID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE config_id = $1`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, configID)
	if err != nil {
		return fmt.Errorf("failed to delete codebase configuration from PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("codebase configuration with ID %s does not exist", configID)
	}

	return nil
}

// CodebaseConfigExists checks if a codebase configuration exists by ID
func (r *PostgresCodebaseConfigRepository) CodebaseConfigExists(ctx context.Context, configID string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE config_id = $1 LIMIT 1`, r.tableName)

	var exists int
	err := r.db.QueryRowContext(ctx, query, configID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if codebase configuration exists: %w", err)
	}

	return true, nil
}

// ListCodebaseConfigs retrieves codebase configurations with pagination and filtering
func (r *PostgresCodebaseConfigRepository) ListCodebaseConfigs(ctx context.Context, opts ListCodebaseConfigsOptions) ([]*CodebaseConfigRecord, string, error) {
	// Build the base query
	query := fmt.Sprintf(`
		SELECT config_id, name, description, provider, url, default_branch,
			   status, created_at, updated_at, tags, metadata, config
		FROM %s
	`, r.tableName)

	var args []interface{}
	var conditions []string
	argIndex := 1

	// Add provider filtering if provided
	if opts.ProviderFilter != nil {
		conditions = append(conditions, fmt.Sprintf("provider = $%d", argIndex))
		args = append(args, string(*opts.ProviderFilter))
		argIndex++
	}

	// Add tag filtering if provided
	if len(opts.TagFilter) > 0 {
		tagConditions := make([]string, 0, len(opts.TagFilter))
		for key, value := range opts.TagFilter {
			tagConditions = append(tagConditions, fmt.Sprintf("tags ->> '%s' = $%d", key, argIndex))
			args = append(args, value)
			argIndex++
		}
		if len(tagConditions) > 0 {
			conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(tagConditions, " AND ")))
		}
	}

	// Add pagination if provided
	if opts.NextToken != nil && *opts.NextToken != "" {
		conditions = append(conditions, fmt.Sprintf("config_id > $%d", argIndex))
		args = append(args, *opts.NextToken)
		argIndex++
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ordering and limit
	query += " ORDER BY config_id"
	if opts.MaxResults != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *opts.MaxResults+1) // Get one extra to check if there are more results
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list codebase configurations from PostgreSQL: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Error closing rows: %v\n", closeErr)
		}
	}()

	var configs []*CodebaseConfigRecord

	for rows.Next() {
		var config CodebaseConfigRecord
		var description sql.NullString
		var tagsJSON, metadataJSON, configJSON []byte

		err := rows.Scan(
			&config.ConfigID,
			&config.Name,
			&description,
			&config.Provider,
			&config.URL,
			&config.DefaultBranch,
			&config.Status,
			&config.CreatedAt,
			&config.UpdatedAt,
			&tagsJSON,
			&metadataJSON,
			&configJSON,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan codebase configuration row: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			config.Description = &description.String
		}

		// Unmarshal JSON fields
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &config.Tags); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &config.Metadata); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &config.Config); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		configs = append(configs, &config)
	}

	// Handle pagination
	var nextToken string
	if opts.MaxResults != nil && len(configs) > *opts.MaxResults {
		// Remove the extra configuration and set the next token
		configs = configs[:*opts.MaxResults]
		nextToken = configs[len(configs)-1].ConfigID
	}

	return configs, nextToken, nil
}

// CreateTable creates the codebase_configs table if it doesn't exist
func (r *PostgresCodebaseConfigRepository) CreateTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			config_id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			provider VARCHAR(50) NOT NULL,
			url VARCHAR(2048) NOT NULL,
			default_branch VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			tags JSONB DEFAULT '{}',
			metadata JSONB DEFAULT '{}',
			config JSONB NOT NULL
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create codebase_configs table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_name ON %s (name)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_provider ON %s (provider)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_status ON %s (status)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s (created_at)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_tags ON %s USING GIN (tags)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_url ON %s (url)", r.tableName, r.tableName),
	}

	for _, indexQuery := range indexes {
		_, err := r.db.ExecContext(ctx, indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
