// Package repository provides PostgreSQL implementation for project data operations
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	conf "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/lib/pq"
)

// PostgresProjectRepository implements ProjectRepository using PostgreSQL
type PostgresProjectRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresProjectRepository creates a new PostgreSQL project repository
func NewPostgresProjectRepository(config PostgresConfig, tableName string) (ProjectRepository, error) {
	if tableName == "" {
		tableName = conf.DefaultProjectsTableName
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

	repo := &PostgresProjectRepository{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := repo.CreateTable(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create projects table: %w", err)
	}

	return repo, nil
}

// NewPostgresProjectRepositoryWithDB creates a new PostgreSQL project repository with an existing DB connection
// This is primarily used for testing with mock databases
func NewPostgresProjectRepositoryWithDB(db *sql.DB, tableName string) ProjectRepository {
	if tableName == "" {
		tableName = "projects"
	}

	return &PostgresProjectRepository{
		db:        db,
		tableName: tableName,
	}
}

// Close closes the database connection
func (r *PostgresProjectRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateProject creates a new project record in PostgreSQL
func (r *PostgresProjectRepository) CreateProject(ctx context.Context, project *ProjectRecord) error {
	// Convert maps to JSON for storage
	tagsJSON, err := json.Marshal(project.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			project_id, name, description, language, status, 
			created_at, updated_at, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, r.tableName)

	_, err = r.db.ExecContext(ctx, query,
		project.ProjectID,
		project.Name,
		project.Description,
		project.Language,
		project.Status,
		project.CreatedAt,
		project.UpdatedAt,
		tagsJSON,
		metadataJSON,
	)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("project with ID %s already exists", project.ProjectID)
		}
		return fmt.Errorf("failed to create project in PostgreSQL: %w", err)
	}

	return nil
}

// GetProject retrieves a project by ID from PostgreSQL
func (r *PostgresProjectRepository) GetProject(ctx context.Context, projectID string) (*ProjectRecord, error) {
	query := fmt.Sprintf(`
		SELECT project_id, name, description, language, status,
			   created_at, updated_at, tags, metadata
		FROM %s WHERE project_id = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, projectID)

	var project ProjectRecord
	var description, language sql.NullString
	var tagsJSON, metadataJSON []byte

	err := row.Scan(
		&project.ProjectID,
		&project.Name,
		&description,
		&language,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
		&tagsJSON,
		&metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Project not found
		}
		return nil, fmt.Errorf("failed to get project from PostgreSQL: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		project.Description = &description.String
	}
	if language.Valid {
		project.Language = &language.String
	}

	// Unmarshal JSON fields
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &project.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &project.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &project, nil
}

// UpdateProject updates an existing project record in PostgreSQL
func (r *PostgresProjectRepository) UpdateProject(ctx context.Context, project *ProjectRecord) error {
	// Convert maps to JSON for storage
	tagsJSON, err := json.Marshal(project.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s SET 
			name = $2, description = $3, language = $4, status = $5,
			updated_at = $6, tags = $7, metadata = $8
		WHERE project_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query,
		project.ProjectID,
		project.Name,
		project.Description,
		project.Language,
		project.Status,
		project.UpdatedAt,
		tagsJSON,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to update project in PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project with ID %s does not exist", project.ProjectID)
	}

	return nil
}

// DeleteProject deletes a project by ID from PostgreSQL
func (r *PostgresProjectRepository) DeleteProject(ctx context.Context, projectID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE project_id = $1`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project from PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project with ID %s does not exist", projectID)
	}

	return nil
}

// ProjectExists checks if a project exists by ID
func (r *PostgresProjectRepository) ProjectExists(ctx context.Context, projectID string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE project_id = $1 LIMIT 1`, r.tableName)

	var exists int
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return true, nil
}

// ListProjects retrieves projects with pagination and filtering
func (r *PostgresProjectRepository) ListProjects(ctx context.Context, opts ListProjectsOptions) ([]*ProjectRecord, string, error) {
	// Build the base query
	query := fmt.Sprintf(`
		SELECT project_id, name, description, language, status,
			   created_at, updated_at, tags, metadata
		FROM %s
	`, r.tableName)

	var args []interface{}
	var conditions []string
	argIndex := 1

	// Add tag filtering if provided
	if len(opts.TagFilter) > 0 {
		for key, value := range opts.TagFilter {
			condition := fmt.Sprintf("tags::jsonb @> $%d::jsonb", argIndex)
			conditions = append(conditions, condition)

			tagFilter := map[string]string{key: value}
			tagJSON, err := json.Marshal(tagFilter)
			if err != nil {
				return nil, "", fmt.Errorf("failed to marshal tag filter: %w", err)
			}
			args = append(args, string(tagJSON))
			argIndex++
		}
	}

	// Add pagination if provided
	if opts.NextToken != nil && *opts.NextToken != "" {
		condition := fmt.Sprintf("project_id > $%d", argIndex)
		conditions = append(conditions, condition)
		args = append(args, *opts.NextToken)
		argIndex++
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	// Add ordering and limit
	query += " ORDER BY project_id"
	if opts.MaxResults != nil {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, *opts.MaxResults+1) // Get one extra to check if there are more results
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list projects: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log the error but don't override the main error
			fmt.Printf("Warning: failed to close rows: %v\n", closeErr)
		}
	}()

	var projects []*ProjectRecord

	for rows.Next() {
		var project ProjectRecord
		var description, language sql.NullString
		var tagsJSON, metadataJSON []byte

		err := rows.Scan(
			&project.ProjectID,
			&project.Name,
			&description,
			&language,
			&project.Status,
			&project.CreatedAt,
			&project.UpdatedAt,
			&tagsJSON,
			&metadataJSON,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan project row: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			project.Description = &description.String
		}
		if language.Valid {
			project.Language = &language.String
		}

		// Unmarshal JSON fields
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &project.Tags); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &project.Metadata); err != nil {
				return nil, "", fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		projects = append(projects, &project)
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate over project rows: %w", err)
	}

	// Handle pagination
	var nextToken string
	if opts.MaxResults != nil && len(projects) > *opts.MaxResults {
		// Remove the extra project and set the next token
		projects = projects[:*opts.MaxResults]
		if len(projects) > 0 {
			nextToken = projects[len(projects)-1].ProjectID
		}
	}

	return projects, nextToken, nil
}

// CreateTable creates the projects table if it doesn't exist
func (r *PostgresProjectRepository) CreateTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			project_id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			language VARCHAR(50),
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			tags JSONB DEFAULT '{}',
			metadata JSONB DEFAULT '{}'
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create projects table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_name ON %s (name)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_status ON %s (status)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s (created_at)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_tags ON %s USING GIN (tags)", r.tableName, r.tableName),
	}

	for _, indexQuery := range indexes {
		if _, err := r.db.ExecContext(ctx, indexQuery); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}
