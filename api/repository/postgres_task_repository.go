package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	conf "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// PostgresTaskRepository implements TaskRepository using PostgreSQL
type PostgresTaskRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresTaskRepository creates a new PostgreSQL task repository
func NewPostgresTaskRepository(config PostgresConfig, tableName string) (TaskRepository, error) {
	if tableName == "" {
		tableName = conf.DefaultTasksTableName
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &PostgresTaskRepository{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := repo.createTableIfNotExists(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return repo, nil
}

// createTableIfNotExists creates the tasks table if it doesn't exist
func (r *PostgresTaskRepository) createTableIfNotExists() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			task_id VARCHAR(255) PRIMARY KEY,
			project_id VARCHAR(255) NOT NULL,
			agent_id VARCHAR(255) NOT NULL,
			codebase_id VARCHAR(255),
			type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			title VARCHAR(500) NOT NULL,
			description TEXT NOT NULL,
			input JSONB,
			output JSONB,
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE,
			metadata JSONB,
			tags JSONB,
			
			-- Indexes for performance
			CONSTRAINT tasks_type_check CHECK (type IN ('code_analysis', 'refactoring', 'code_review', 'documentation', 'custom')),
			CONSTRAINT tasks_status_check CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'cancelled'))
		);
		
		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_%s_project_id ON %s (project_id);
		CREATE INDEX IF NOT EXISTS idx_%s_agent_id ON %s (agent_id);
		CREATE INDEX IF NOT EXISTS idx_%s_codebase_id ON %s (codebase_id);
		CREATE INDEX IF NOT EXISTS idx_%s_status ON %s (status);
		CREATE INDEX IF NOT EXISTS idx_%s_type ON %s (type);
		CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s (created_at);
		CREATE INDEX IF NOT EXISTS idx_%s_project_status ON %s (project_id, status);
	`, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName,
		r.tableName, r.tableName, r.tableName, r.tableName, r.tableName, r.tableName,
		r.tableName, r.tableName, r.tableName)

	_, err := r.db.Exec(query)
	return err
}

// Create creates a new task
func (r *PostgresTaskRepository) Create(ctx context.Context, task *models.Task) error {
	// Generate UUID if not provided
	if task.TaskID == "" {
		task.TaskID = "task-" + uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Convert maps to JSON
	inputJSON, _ := json.Marshal(task.Input)
	outputJSON, _ := json.Marshal(task.Output)
	metadataJSON, _ := json.Marshal(task.Metadata)
	tagsJSON, _ := json.Marshal(task.Tags)

	query := fmt.Sprintf(`
		INSERT INTO %s (
			task_id, project_id, agent_id, codebase_id, type, status, 
			title, description, input, output, error_message,
			created_at, updated_at, completed_at, metadata, tags
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query,
		task.TaskID, task.ProjectID, task.AgentID, task.CodebaseID, task.Type, task.Status,
		task.Title, task.Description, inputJSON, outputJSON, task.ErrorMessage,
		task.CreatedAt, task.UpdatedAt, task.CompletedAt, metadataJSON, tagsJSON,
	)

	return err
}

// GetByID retrieves a task by its ID
func (r *PostgresTaskRepository) GetByID(ctx context.Context, taskID string) (*models.Task, error) {
	query := fmt.Sprintf(`
		SELECT task_id, project_id, agent_id, codebase_id, type, status,
			   title, description, input, output, error_message,
			   created_at, updated_at, completed_at, metadata, tags
		FROM %s
		WHERE task_id = $1
	`, r.tableName)

	var task models.Task
	var inputJSON, outputJSON, metadataJSON, tagsJSON []byte

	err := r.db.QueryRowContext(ctx, query, taskID).Scan(
		&task.TaskID, &task.ProjectID, &task.AgentID, &task.CodebaseID, &task.Type, &task.Status,
		&task.Title, &task.Description, &inputJSON, &outputJSON, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &metadataJSON, &tagsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, err
	}

	// Parse JSON fields
	if len(inputJSON) > 0 {
		if err := json.Unmarshal(inputJSON, &task.Input); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input JSON: %w", err)
		}
	}
	if len(outputJSON) > 0 {
		if err := json.Unmarshal(outputJSON, &task.Output); err != nil {
			return nil, fmt.Errorf("failed to unmarshal output JSON: %w", err)
		}
	}
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata JSON: %w", err)
		}
	}
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &task.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags JSON: %w", err)
		}
	}

	return &task, nil
}

// Update updates an existing task
func (r *PostgresTaskRepository) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()

	// Convert maps to JSON
	inputJSON, _ := json.Marshal(task.Input)
	outputJSON, _ := json.Marshal(task.Output)
	metadataJSON, _ := json.Marshal(task.Metadata)
	tagsJSON, _ := json.Marshal(task.Tags)

	query := fmt.Sprintf(`
		UPDATE %s SET
			project_id = $2, agent_id = $3, codebase_id = $4, type = $5, status = $6,
			title = $7, description = $8, input = $9, output = $10, error_message = $11,
			updated_at = $12, completed_at = $13, metadata = $14, tags = $15
		WHERE task_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query,
		task.TaskID, task.ProjectID, task.AgentID, task.CodebaseID, task.Type, task.Status,
		task.Title, task.Description, inputJSON, outputJSON, task.ErrorMessage,
		task.UpdatedAt, task.CompletedAt, metadataJSON, tagsJSON,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", task.TaskID)
	}

	return nil
}

// Delete deletes a task by its ID
func (r *PostgresTaskRepository) Delete(ctx context.Context, taskID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE task_id = $1`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// ListByProject lists tasks for a specific project with optional filters
func (r *PostgresTaskRepository) ListByProject(ctx context.Context, projectID string, filters TaskFilters) ([]models.Task, int, error) {
	whereClause := "WHERE project_id = $1"
	args := []interface{}{projectID}
	argIndex := 2

	// Add filters
	if filters.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}
	if filters.Type != nil {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filters.Type)
		argIndex++
	}
	if filters.AgentID != nil {
		whereClause += fmt.Sprintf(" AND agent_id = $%d", argIndex)
		args = append(args, *filters.AgentID)
		argIndex++
	}
	if filters.CodebaseID != nil {
		whereClause += fmt.Sprintf(" AND codebase_id = $%d", argIndex)
		args = append(args, *filters.CodebaseID)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.tableName, whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT task_id, project_id, agent_id, codebase_id, type, status,
			   title, description, input, output, error_message,
			   created_at, updated_at, completed_at, metadata, tags
		FROM %s %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, r.tableName, whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Warn("failed to close rows in ListByProject", "error", closeErr)
		}
	}()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var inputJSON, outputJSON, metadataJSON, tagsJSON []byte

		err := rows.Scan(
			&task.TaskID, &task.ProjectID, &task.AgentID, &task.CodebaseID, &task.Type, &task.Status,
			&task.Title, &task.Description, &inputJSON, &outputJSON, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &metadataJSON, &tagsJSON,
		)
		if err != nil {
			return nil, 0, err
		}

		// Parse JSON fields with error handling
		if len(inputJSON) > 0 {
			if err := json.Unmarshal(inputJSON, &task.Input); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal input JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(outputJSON) > 0 {
			if err := json.Unmarshal(outputJSON, &task.Output); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal output JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &task.Tags); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal tags JSON for task %s: %w", task.TaskID, err)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, totalCount, rows.Err()
}

// ListByAgent lists tasks for a specific agent
func (r *PostgresTaskRepository) ListByAgent(ctx context.Context, agentID string, filters TaskFilters) ([]models.Task, int, error) {
	filters.AgentID = &agentID
	return r.listWithFilters(ctx, filters)
}

// ListByCodebase lists tasks for a specific codebase
func (r *PostgresTaskRepository) ListByCodebase(ctx context.Context, codebaseID string, filters TaskFilters) ([]models.Task, int, error) {
	filters.CodebaseID = &codebaseID
	return r.listWithFilters(ctx, filters)
}

// UpdateStatus updates only the status of a task
func (r *PostgresTaskRepository) UpdateStatus(ctx context.Context, taskID string, status models.TaskStatus) error {
	now := time.Now()
	var completedAt *time.Time

	// Set completed_at if status is completed
	if status == models.TaskStatusCompleted {
		completedAt = &now
	}

	query := fmt.Sprintf(`
		UPDATE %s SET status = $2, updated_at = $3, completed_at = $4
		WHERE task_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, taskID, status, now, completedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// UpdateStatusAndOutput updates the status and output of a task
func (r *PostgresTaskRepository) UpdateStatusAndOutput(ctx context.Context, taskID string, status models.TaskStatus, output map[string]any, errorMessage *string) error {
	now := time.Now()
	var completedAt *time.Time

	// Set completed_at if status is completed or failed
	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
		completedAt = &now
	}

	outputJSON, _ := json.Marshal(output)

	query := fmt.Sprintf(`
		UPDATE %s SET status = $2, output = $3, error_message = $4, updated_at = $5, completed_at = $6
		WHERE task_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, taskID, status, outputJSON, errorMessage, now, completedAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	return nil
}

// listWithFilters is a helper method for listing tasks with filters
func (r *PostgresTaskRepository) listWithFilters(ctx context.Context, filters TaskFilters) ([]models.Task, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add filters
	if filters.Status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}
	if filters.Type != nil {
		whereClause += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, *filters.Type)
		argIndex++
	}
	if filters.AgentID != nil {
		whereClause += fmt.Sprintf(" AND agent_id = $%d", argIndex)
		args = append(args, *filters.AgentID)
		argIndex++
	}
	if filters.CodebaseID != nil {
		whereClause += fmt.Sprintf(" AND codebase_id = $%d", argIndex)
		args = append(args, *filters.CodebaseID)
		argIndex++
	}

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, r.tableName, whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Main query with pagination
	query := fmt.Sprintf(`
		SELECT task_id, project_id, agent_id, codebase_id, type, status,
			   title, description, input, output, error_message,
			   created_at, updated_at, completed_at, metadata, tags
		FROM %s %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, r.tableName, whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.Warn("failed to close rows in listWithFilters", "error", closeErr)
		}
	}()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		var inputJSON, outputJSON, metadataJSON, tagsJSON []byte

		err := rows.Scan(
			&task.TaskID, &task.ProjectID, &task.AgentID, &task.CodebaseID, &task.Type, &task.Status,
			&task.Title, &task.Description, &inputJSON, &outputJSON, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt, &task.CompletedAt, &metadataJSON, &tagsJSON,
		)
		if err != nil {
			return nil, 0, err
		}

		// Parse JSON fields with error handling
		if len(inputJSON) > 0 {
			if err := json.Unmarshal(inputJSON, &task.Input); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal input JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(outputJSON) > 0 {
			if err := json.Unmarshal(outputJSON, &task.Output); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal output JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata JSON for task %s: %w", task.TaskID, err)
			}
		}
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &task.Tags); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal tags JSON for task %s: %w", task.TaskID, err)
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, totalCount, rows.Err()
}
