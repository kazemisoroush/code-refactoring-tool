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

func TestNewPostgresProjectRepository(t *testing.T) {
	// This test requires a real PostgreSQL connection, so we'll skip it in CI
	t.Skip("Skipping test that requires real PostgreSQL connection")
}

func TestNewPostgresProjectRepositoryWithDB(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "test_projects")
	assert.NotNil(t, repo)

	pgRepo := repo.(*PostgresProjectRepository)
	assert.Equal(t, "test_projects", pgRepo.tableName)
}

func TestPostgresProjectRepository_CreateProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	project := &ProjectRecord{
		ProjectID:   "proj-12345",
		Name:        "test-project",
		Description: stringPtr("A test project"),
		Language:    stringPtr("go"),
		Status:      string(models.ProjectStatusActive),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Tags: map[string]string{
			"env": "test",
		},
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	mock.ExpectExec(`INSERT INTO projects`).
		WithArgs(
			project.ProjectID,
			project.Name,
			project.Description,
			project.Language,
			project.Status,
			project.CreatedAt,
			project.UpdatedAt,
			sqlmock.AnyArg(), // tags JSON
			sqlmock.AnyArg(), // metadata JSON
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateProject(context.Background(), project)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_CreateProject_DuplicateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	project := &ProjectRecord{
		ProjectID:   "proj-12345",
		Name:        "Test Project",
		Description: stringPtr("Test Description"),
		Language:    stringPtr("Go"),
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        map[string]string{"env": "test"},
		Metadata:    map[string]string{"version": "1.0"},
	}

	// Create a real pq.Error for unique constraint violation
	pqErr := &pq.Error{
		Code: "23505", // unique_violation
	}

	mock.ExpectExec("INSERT INTO projects").
		WithArgs(
			project.ProjectID,
			project.Name,
			project.Description,
			project.Language,
			project.Status,
			project.CreatedAt,
			project.UpdatedAt,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(pqErr)

	err = repo.CreateProject(context.Background(), project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_GetProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "proj-12345"
	description := "A test project"
	language := "go"
	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()
	tagsJSON := `{"env":"test"}`
	metadataJSON := `{"version":"1.0.0"}`

	rows := sqlmock.NewRows([]string{
		"project_id", "name", "description", "language", "status",
		"created_at", "updated_at", "tags", "metadata",
	}).AddRow(
		projectID, "test-project", description, language, "active",
		createdAt, updatedAt, []byte(tagsJSON), []byte(metadataJSON),
	)

	mock.ExpectQuery(`SELECT (.+) FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnRows(rows)

	project, err := repo.GetProject(context.Background(), projectID)
	require.NoError(t, err)
	require.NotNil(t, project)

	assert.Equal(t, projectID, project.ProjectID)
	assert.Equal(t, "test-project", project.Name)
	assert.Equal(t, &description, project.Description)
	assert.Equal(t, &language, project.Language)
	assert.Equal(t, "active", project.Status)
	assert.Equal(t, "test", project.Tags["env"])
	assert.Equal(t, "1.0.0", project.Metadata["version"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_GetProject_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "nonexistent"

	mock.ExpectQuery(`SELECT (.+) FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnError(sql.ErrNoRows)

	project, err := repo.GetProject(context.Background(), projectID)
	assert.NoError(t, err)
	assert.Nil(t, project)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_UpdateProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	project := &ProjectRecord{
		ProjectID:   "proj-12345",
		Name:        "updated-project",
		Description: stringPtr("Updated description"),
		Language:    stringPtr("python"),
		Status:      string(models.ProjectStatusActive),
		UpdatedAt:   time.Now().UTC(),
		Tags: map[string]string{
			"env": "staging",
		},
		Metadata: map[string]string{
			"version": "1.1.0",
		},
	}

	mock.ExpectExec(`UPDATE projects SET`).
		WithArgs(
			project.ProjectID,
			project.Name,
			project.Description,
			project.Language,
			project.Status,
			project.UpdatedAt,
			sqlmock.AnyArg(), // tags JSON
			sqlmock.AnyArg(), // metadata JSON
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UpdateProject(context.Background(), project)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_UpdateProject_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	project := &ProjectRecord{
		ProjectID: "nonexistent",
		Name:      "updated-project",
		Status:    string(models.ProjectStatusActive),
		UpdatedAt: time.Now().UTC(),
		Tags:      make(map[string]string),
		Metadata:  make(map[string]string),
	}

	mock.ExpectExec(`UPDATE projects SET`).
		WithArgs(
			project.ProjectID,
			project.Name,
			project.Description,
			project.Language,
			project.Status,
			project.UpdatedAt,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	err = repo.UpdateProject(context.Background(), project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_DeleteProject_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "proj-12345"

	mock.ExpectExec(`DELETE FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteProject(context.Background(), projectID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_DeleteProject_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "nonexistent"

	mock.ExpectExec(`DELETE FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	err = repo.DeleteProject(context.Background(), projectID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_ProjectExists_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "proj-12345"

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(1)
	mock.ExpectQuery(`SELECT 1 FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnRows(rows)

	exists, err := repo.ProjectExists(context.Background(), projectID)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_ProjectExists_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	projectID := "nonexistent"

	mock.ExpectQuery(`SELECT 1 FROM projects WHERE project_id`).
		WithArgs(projectID).
		WillReturnError(sql.ErrNoRows)

	exists, err := repo.ProjectExists(context.Background(), projectID)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_ListProjects_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	maxResults := 2
	opts := ListProjectsOptions{
		MaxResults: &maxResults,
	}

	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"project_id", "name", "description", "language", "status",
		"created_at", "updated_at", "tags", "metadata",
	}).
		AddRow("proj-12345", "project-1", "desc-1", "go", "active",
			createdAt, updatedAt, []byte(`{"env":"test"}`), []byte(`{"version":"1.0.0"}`)).
		AddRow("proj-67890", "project-2", "desc-2", "python", "active",
			createdAt, updatedAt, []byte(`{"env":"prod"}`), []byte(`{"version":"2.0.0"}`))

	mock.ExpectQuery(`SELECT (.+) FROM projects ORDER BY project_id LIMIT`).
		WithArgs(3). // maxResults + 1
		WillReturnRows(rows)

	projects, nextToken, err := repo.ListProjects(context.Background(), opts)
	require.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Empty(t, nextToken) // No next token since we didn't hit the limit

	assert.Equal(t, "proj-12345", projects[0].ProjectID)
	assert.Equal(t, "project-1", projects[0].Name)
	assert.Equal(t, "proj-67890", projects[1].ProjectID)
	assert.Equal(t, "project-2", projects[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_ListProjects_WithTagFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	opts := ListProjectsOptions{
		TagFilter: map[string]string{
			"env": "test",
		},
	}

	createdAt := time.Now().UTC()
	updatedAt := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"project_id", "name", "description", "language", "status",
		"created_at", "updated_at", "tags", "metadata",
	}).
		AddRow("proj-12345", "project-1", "desc-1", "go", "active",
			createdAt, updatedAt, []byte(`{"env":"test"}`), []byte(`{"version":"1.0.0"}`))

	mock.ExpectQuery(`SELECT (.+) FROM projects WHERE tags::jsonb @> (.+) ORDER BY project_id`).
		WithArgs(`{"env":"test"}`).
		WillReturnRows(rows)

	projects, nextToken, err := repo.ListProjects(context.Background(), opts)
	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.Empty(t, nextToken)

	assert.Equal(t, "proj-12345", projects[0].ProjectID)
	assert.Equal(t, "test", projects[0].Tags["env"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresProjectRepository_CreateTable_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck // Test cleanup

	repo := NewPostgresProjectRepositoryWithDB(db, "projects")

	// Expect table creation
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS projects`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Expect index creation
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_projects_name`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_projects_status`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_projects_created_at`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`CREATE INDEX IF NOT EXISTS idx_projects_tags`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.(*PostgresProjectRepository).CreateTable(context.Background())
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Helper functions for tests

func stringPtr(s string) *string {
	return &s
}
