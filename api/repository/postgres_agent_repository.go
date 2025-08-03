// Package repository provides PostgreSQL implementation for agent data operations
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	conf "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// PostgresConfig holds configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// PostgresAgentRepository implements AgentRepository using PostgreSQL
type PostgresAgentRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresAgentRepository creates a new PostgreSQL agent repository
func NewPostgresAgentRepository(config PostgresConfig, tableName string) (AgentRepository, error) {
	if tableName == "" {
		tableName = conf.DefaultAgentsTableName
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

	return &PostgresAgentRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// NewPostgresAgentRepositoryWithDB creates a new PostgreSQL agent repository with an existing DB connection
// This is primarily used for testing with mock databases
func NewPostgresAgentRepositoryWithDB(db *sql.DB, tableName string) AgentRepository {
	if tableName == "" {
		tableName = "agents"
	}

	return &PostgresAgentRepository{
		db:        db,
		tableName: tableName,
	}
}

// Close closes the database connection
func (r *PostgresAgentRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// CreateAgent stores a new agent record
func (r *PostgresAgentRepository) CreateAgent(ctx context.Context, agent *AgentRecord) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			agent_id, agent_version, knowledge_base_id, vector_store_id,
			repository_url, branch, agent_name, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query,
		agent.AgentID,
		agent.AgentVersion,
		agent.KnowledgeBaseID,
		agent.VectorStoreID,
		agent.RepositoryURL,
		agent.Branch,
		agent.AgentName,
		agent.Status,
		agent.CreatedAt,
		agent.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return fmt.Errorf("agent with ID %s already exists", agent.AgentID)
		}
		return fmt.Errorf("failed to create agent in PostgreSQL: %w", err)
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (r *PostgresAgentRepository) GetAgent(ctx context.Context, agentID string) (*AgentRecord, error) {
	query := fmt.Sprintf(`
		SELECT agent_id, agent_version, knowledge_base_id, vector_store_id,
			   repository_url, branch, agent_name, status, created_at, updated_at
		FROM %s WHERE agent_id = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, agentID)

	var agent AgentRecord
	var branch, agentName sql.NullString

	err := row.Scan(
		&agent.AgentID,
		&agent.AgentVersion,
		&agent.KnowledgeBaseID,
		&agent.VectorStoreID,
		&agent.RepositoryURL,
		&branch,
		&agentName,
		&agent.Status,
		&agent.CreatedAt,
		&agent.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
		return nil, fmt.Errorf("failed to get agent from PostgreSQL: %w", err)
	}

	// Handle nullable fields
	if branch.Valid {
		agent.Branch = branch.String
	}
	if agentName.Valid {
		agent.AgentName = agentName.String
	}

	return &agent, nil
}

// UpdateAgent updates an existing agent record
func (r *PostgresAgentRepository) UpdateAgent(ctx context.Context, agent *AgentRecord) error {
	agent.UpdatedAt = time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s SET
			agent_version = $2,
			knowledge_base_id = $3,
			vector_store_id = $4,
			repository_url = $5,
			branch = $6,
			agent_name = $7,
			status = $8,
			updated_at = $9
		WHERE agent_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query,
		agent.AgentID,
		agent.AgentVersion,
		agent.KnowledgeBaseID,
		agent.VectorStoreID,
		agent.RepositoryURL,
		agent.Branch,
		agent.AgentName,
		agent.Status,
		agent.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent in PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", agent.AgentID)
	}

	return nil
}

// DeleteAgent removes an agent record
func (r *PostgresAgentRepository) DeleteAgent(ctx context.Context, agentID string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE agent_id = $1", r.tableName)

	result, err := r.db.ExecContext(ctx, query, agentID)
	if err != nil {
		return fmt.Errorf("failed to delete agent from PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return nil
}

// ListAgents retrieves all agent records
func (r *PostgresAgentRepository) ListAgents(ctx context.Context) ([]*AgentRecord, error) {
	query := fmt.Sprintf(`
		SELECT agent_id, agent_version, knowledge_base_id, vector_store_id,
			   repository_url, branch, agent_name, status, created_at, updated_at
		FROM %s ORDER BY created_at DESC
	`, r.tableName)

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents from PostgreSQL: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("Warning: failed to close rows: %v", closeErr)
		}
	}()

	var agents []*AgentRecord
	for rows.Next() {
		var agent AgentRecord
		var branch, agentName sql.NullString

		err := rows.Scan(
			&agent.AgentID,
			&agent.AgentVersion,
			&agent.KnowledgeBaseID,
			&agent.VectorStoreID,
			&agent.RepositoryURL,
			&branch,
			&agentName,
			&agent.Status,
			&agent.CreatedAt,
			&agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent record: %w", err)
		}

		// Handle nullable fields
		if branch.Valid {
			agent.Branch = branch.String
		}
		if agentName.Valid {
			agent.AgentName = agentName.String
		}

		agents = append(agents, &agent)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over agent records: %w", err)
	}

	return agents, nil
}

// UpdateAgentStatus updates only the status field
func (r *PostgresAgentRepository) UpdateAgentStatus(ctx context.Context, agentID string, status models.AgentStatus) error {
	query := fmt.Sprintf(`
		UPDATE %s SET status = $2, updated_at = $3 WHERE agent_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, agentID, string(status), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update agent status in PostgreSQL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return nil
}

// CreateAgentsTable creates the agents table if it doesn't exist
func (r *PostgresAgentRepository) CreateAgentsTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			agent_id VARCHAR(255) PRIMARY KEY,
			agent_version VARCHAR(255) NOT NULL,
			knowledge_base_id VARCHAR(255) NOT NULL,
			vector_store_id VARCHAR(255) NOT NULL,
			repository_url TEXT NOT NULL,
			branch VARCHAR(255),
			agent_name VARCHAR(255),
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create agents table: %w", err)
	}

	// Create index on status for efficient filtering
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_status ON %s(status)
	`, r.tableName, r.tableName)

	_, err = r.db.ExecContext(ctx, indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	// Create index on created_at for efficient sorting
	timeIndexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s(created_at DESC)
	`, r.tableName, r.tableName)

	_, err = r.db.ExecContext(ctx, timeIndexQuery)
	if err != nil {
		return fmt.Errorf("failed to create created_at index: %w", err)
	}

	return nil
}
