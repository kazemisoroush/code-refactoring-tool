package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

func TestNewPostgresAgentRepository(t *testing.T) {
	tests := []struct {
		name          string
		config        PostgresConfig
		tableName     string
		expectedTable string
		expectError   bool
	}{
		{
			name: "with valid config and custom table name",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "disable",
			},
			tableName:     "custom_agents",
			expectedTable: "custom_agents",
			expectError:   true, // Will fail in test environment without real DB
		},
		{
			name: "with valid config and empty table name uses default",
			config: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "testuser",
				Password: "testpass",
				SSLMode:  "disable",
			},
			tableName:     "",
			expectedTable: "agents",
			expectError:   true, // Will fail in test environment without real DB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewPostgresAgentRepository(tt.config, tt.tableName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				defer func() {
					if closer, ok := repo.(*PostgresAgentRepository); ok {
						_ = closer.Close()
					}
				}()
			}
		})
	}
}

func TestNewPostgresAgentRepositoryWithDB(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	tests := []struct {
		name          string
		tableName     string
		expectedTable string
	}{
		{
			name:          "with custom table name",
			tableName:     "custom_agents",
			expectedTable: "custom_agents",
		},
		{
			name:          "with empty table name uses default",
			tableName:     "",
			expectedTable: "agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewPostgresAgentRepositoryWithDB(db, tt.tableName)

			assert.NotNil(t, repo)

			postgresRepo, ok := repo.(*PostgresAgentRepository)
			assert.True(t, ok, "Repository should be of type *PostgresAgentRepository")
			assert.Equal(t, tt.expectedTable, postgresRepo.tableName)
			assert.NotNil(t, postgresRepo.db)
		})
	}
}

func TestPostgresAgentRepository_CreateAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()

	now := time.Now().UTC()
	agent := &AgentRecord{
		AgentID:         "test-agent-id",
		AgentVersion:    "v1.0.0",
		KnowledgeBaseID: "kb-123",
		VectorStoreID:   "vs-456",
		RepositoryURL:   "https://github.com/test/repo",
		Branch:          "main",
		AgentName:       "Test Agent",
		Status:          string(models.AgentStatusReady),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	t.Run("successful creation", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO agents`).
			WithArgs(
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
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateAgent(ctx, agent)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate agent ID", func(t *testing.T) {
		pqErr := &pq.Error{Code: "23505"} // Unique constraint violation
		mock.ExpectExec(`INSERT INTO agents`).
			WithArgs(
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
			).
			WillReturnError(pqErr)

		err := repo.CreateAgent(ctx, agent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_GetAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()

	now := time.Now().UTC()
	agentID := "test-agent-id"

	t.Run("successful retrieval", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"agent_id", "agent_version", "knowledge_base_id", "vector_store_id",
			"repository_url", "branch", "agent_name", "status", "created_at", "updated_at",
		}).AddRow(
			agentID, "v1.0.0", "kb-123", "vs-456",
			"https://github.com/test/repo", "main", "Test Agent", "ready", now, now,
		)

		mock.ExpectQuery(`SELECT .+ FROM agents WHERE agent_id`).
			WithArgs(agentID).
			WillReturnRows(rows)

		agent, err := repo.GetAgent(ctx, agentID)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, agentID, agent.AgentID)
		assert.Equal(t, "v1.0.0", agent.AgentVersion)
		assert.Equal(t, "main", agent.Branch)
		assert.Equal(t, "Test Agent", agent.AgentName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("agent not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT .+ FROM agents WHERE agent_id`).
			WithArgs(agentID).
			WillReturnError(sql.ErrNoRows)

		agent, err := repo.GetAgent(ctx, agentID)
		assert.Error(t, err)
		assert.Nil(t, agent)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("with null optional fields", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"agent_id", "agent_version", "knowledge_base_id", "vector_store_id",
			"repository_url", "branch", "agent_name", "status", "created_at", "updated_at",
		}).AddRow(
			agentID, "v1.0.0", "kb-123", "vs-456",
			"https://github.com/test/repo", nil, nil, "ready", now, now,
		)

		mock.ExpectQuery(`SELECT .+ FROM agents WHERE agent_id`).
			WithArgs(agentID).
			WillReturnRows(rows)

		agent, err := repo.GetAgent(ctx, agentID)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Equal(t, "", agent.Branch)
		assert.Equal(t, "", agent.AgentName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_UpdateAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()

	now := time.Now().UTC()
	agent := &AgentRecord{
		AgentID:         "test-agent-id",
		AgentVersion:    "v1.1.0",
		KnowledgeBaseID: "kb-123",
		VectorStoreID:   "vs-456",
		RepositoryURL:   "https://github.com/test/repo",
		Branch:          "develop",
		AgentName:       "Updated Agent",
		Status:          string(models.AgentStatusError),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	t.Run("successful update", func(t *testing.T) {
		mock.ExpectExec(`UPDATE agents SET`).
			WithArgs(
				agent.AgentID,
				agent.AgentVersion,
				agent.KnowledgeBaseID,
				agent.VectorStoreID,
				agent.RepositoryURL,
				agent.Branch,
				agent.AgentName,
				agent.Status,
				sqlmock.AnyArg(), // updated_at will be set to current time
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateAgent(ctx, agent)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("agent not found", func(t *testing.T) {
		mock.ExpectExec(`UPDATE agents SET`).
			WithArgs(
				agent.AgentID,
				agent.AgentVersion,
				agent.KnowledgeBaseID,
				agent.VectorStoreID,
				agent.RepositoryURL,
				agent.Branch,
				agent.AgentName,
				agent.Status,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateAgent(ctx, agent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_DeleteAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()
	agentID := "test-agent-id"

	t.Run("successful deletion", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM agents WHERE agent_id`).
			WithArgs(agentID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteAgent(ctx, agentID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("agent not found", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM agents WHERE agent_id`).
			WithArgs(agentID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteAgent(ctx, agentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_ListAgents(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()

	now := time.Now().UTC()

	t.Run("successful listing", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"agent_id", "agent_version", "knowledge_base_id", "vector_store_id",
			"repository_url", "branch", "agent_name", "status", "created_at", "updated_at",
		}).
			AddRow("agent-1", "v1.0.0", "kb-1", "vs-1", "https://github.com/test/repo1", "main", "Agent 1", "ready", now, now).
			AddRow("agent-2", "v1.0.0", "kb-2", "vs-2", "https://github.com/test/repo2", nil, nil, "processing", now, now)

		mock.ExpectQuery(`SELECT .+ FROM agents ORDER BY created_at DESC`).
			WillReturnRows(rows)

		agents, err := repo.ListAgents(ctx)
		assert.NoError(t, err)
		assert.Len(t, agents, 2)

		assert.Equal(t, "agent-1", agents[0].AgentID)
		assert.Equal(t, "main", agents[0].Branch)
		assert.Equal(t, "Agent 1", agents[0].AgentName)

		assert.Equal(t, "agent-2", agents[1].AgentID)
		assert.Equal(t, "", agents[1].Branch)
		assert.Equal(t, "", agents[1].AgentName)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"agent_id", "agent_version", "knowledge_base_id", "vector_store_id",
			"repository_url", "branch", "agent_name", "status", "created_at", "updated_at",
		})

		mock.ExpectQuery(`SELECT .+ FROM agents ORDER BY created_at DESC`).
			WillReturnRows(rows)

		agents, err := repo.ListAgents(ctx)
		assert.NoError(t, err)
		assert.Len(t, agents, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_UpdateAgentStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents")
	ctx := context.Background()
	agentID := "test-agent-id"
	status := models.AgentStatusError

	t.Run("successful status update", func(t *testing.T) {
		mock.ExpectExec(`UPDATE agents SET status`).
			WithArgs(agentID, string(status), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateAgentStatus(ctx, agentID, status)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("agent not found", func(t *testing.T) {
		mock.ExpectExec(`UPDATE agents SET status`).
			WithArgs(agentID, string(status), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateAgentStatus(ctx, agentID, status)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPostgresAgentRepository_CreateAgentsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewPostgresAgentRepositoryWithDB(db, "agents").(*PostgresAgentRepository)
	ctx := context.Background()

	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS agents`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_agents_status`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_agents_created_at`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.CreateAgentsTable(ctx)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
