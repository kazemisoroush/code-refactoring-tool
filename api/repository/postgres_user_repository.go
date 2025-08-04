package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db        *sql.DB
	tableName string
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(config PostgresConfig, tableName string) (UserRepository, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &PostgresUserRepository{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := repo.CreateTable(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	return repo, nil
}

// NewPostgresUserRepositoryWithDB creates a new PostgreSQL user repository with existing DB connection
func NewPostgresUserRepositoryWithDB(db *sql.DB, tableName string) UserRepository {
	return &PostgresUserRepository{
		db:        db,
		tableName: tableName,
	}
}

// CreateUser creates a new user in the database
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *models.DBUser) (*models.DBUser, error) {
	// Generate a new user ID if not provided
	if user.UserID == "" {
		user.UserID = generateUserID()
	}

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, auth_id, email, username, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.AuthID,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		string(user.Role),
		string(user.Status),
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			if pqErr.Constraint == "users_email_key" || pqErr.Constraint == "users_username_key" || pqErr.Constraint == "users_auth_id_key" {
				return nil, fmt.Errorf("user with email '%s', username '%s', or auth_id '%s' already exists", user.Email, user.Username, user.AuthID)
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by user ID
func (r *PostgresUserRepository) GetUser(ctx context.Context, userID string) (*models.DBUser, error) {
	query := fmt.Sprintf(`
		SELECT user_id, auth_id, email, username, first_name, last_name, role, status, created_at, updated_at
		FROM %s
		WHERE user_id = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, userID)

	user := &models.DBUser{}
	err := row.Scan(
		&user.UserID,
		&user.AuthID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByAuthID retrieves a user by auth provider ID
func (r *PostgresUserRepository) GetUserByAuthID(ctx context.Context, authID string) (*models.DBUser, error) {
	query := fmt.Sprintf(`
		SELECT user_id, auth_id, email, username, first_name, last_name, role, status, created_at, updated_at
		FROM %s
		WHERE auth_id = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, authID)

	user := &models.DBUser{}
	err := row.Scan(
		&user.UserID,
		&user.AuthID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with auth_id '%s' not found", authID)
		}
		return nil, fmt.Errorf("failed to get user by auth ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *PostgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.DBUser, error) {
	query := fmt.Sprintf(`
		SELECT user_id, auth_id, email, username, first_name, last_name, role, status, created_at, updated_at
		FROM %s
		WHERE email = $1
	`, r.tableName)

	row := r.db.QueryRowContext(ctx, query, email)

	user := &models.DBUser{}
	err := row.Scan(
		&user.UserID,
		&user.AuthID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *models.DBUser) (*models.DBUser, error) {
	user.UpdatedAt = time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s
		SET auth_id = $2, email = $3, username = $4, first_name = $5, last_name = $6, role = $7, status = $8, updated_at = $9
		WHERE user_id = $1
	`, r.tableName)

	result, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.AuthID,
		user.Email,
		user.Username,
		user.FirstName,
		user.LastName,
		string(user.Role),
		string(user.Status),
		user.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return nil, fmt.Errorf("user with email '%s', username '%s', or auth_id '%s' already exists", user.Email, user.Username, user.AuthID)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("user with ID '%s' does not exist", user.UserID)
	}

	return user, nil
}

// DeleteUser deletes a user by user ID
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE user_id = $1`, r.tableName)

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID '%s' does not exist", userID)
	}

	return nil
}

// ListUsers lists users with optional filtering
func (r *PostgresUserRepository) ListUsers(ctx context.Context, filter *ListUsersFilter) ([]*models.DBUser, int, error) {
	whereClause := ""
	args := []interface{}{}
	argCount := 0

	if filter.Role != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND role = $%d", argCount)
		args = append(args, string(*filter.Role))
	}

	if filter.Status != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(*filter.Status))
	}

	if filter.Search != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+filter.Search+"%")
	}

	if whereClause != "" {
		whereClause = "WHERE" + whereClause[4:] // Remove the first " AND"
	}

	// Count total records
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, r.tableName, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	query := fmt.Sprintf(`
		SELECT user_id, auth_id, email, username, first_name, last_name, role, status, created_at, updated_at
		FROM %s
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, r.tableName, whereClause, argCount+1, argCount+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer func() {
		_ = rows.Close() // Ignore close error as it's in defer
	}()

	var users []*models.DBUser
	for rows.Next() {
		user := &models.DBUser{}
		err := rows.Scan(
			&user.UserID,
			&user.AuthID,
			&user.Email,
			&user.Username,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

// UserExists checks if a user exists by user ID
func (r *PostgresUserRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE user_id = $1`, r.tableName)

	var exists int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return true, nil
}

// HasProjectAccess checks if a user has access to a project
func (r *PostgresUserRepository) HasProjectAccess(ctx context.Context, userID, _ /* projectID */, accessType string) (bool, error) {
	// For now, we'll implement a simple role-based access control
	// In the future, this could be extended to support project-specific permissions
	// TODO: Use projectID parameter when implementing project-specific access control

	user, err := r.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	// Owners and Admins have access to everything
	if user.Role == models.RoleOwner || user.Role == models.RoleAdmin {
		return true, nil
	}

	// Developers have read/write access to projects
	if user.Role == models.RoleDeveloper && (accessType == "read" || accessType == "write") {
		return true, nil
	}

	// Viewers have read-only access
	if user.Role == models.RoleViewer && accessType == "read" {
		return true, nil
	}

	return false, nil
}

// GrantProjectAccess grants a user access to a project (placeholder for future implementation)
func (r *PostgresUserRepository) GrantProjectAccess(_ context.Context, _, _, _ string) error {
	// This is a placeholder implementation
	// In the future, this could create entries in a user_project_access table
	return fmt.Errorf("project access management not yet implemented")
}

// RevokeProjectAccess revokes a user's access to a project (placeholder for future implementation)
func (r *PostgresUserRepository) RevokeProjectAccess(_ context.Context, _, _ string) error {
	// This is a placeholder implementation
	// In the future, this could remove entries from a user_project_access table
	return fmt.Errorf("project access management not yet implemented")
}

// CreateTable creates the users table with appropriate indexes
func (r *PostgresUserRepository) CreateTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			user_id VARCHAR(255) PRIMARY KEY,
			auth_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			role VARCHAR(50) NOT NULL DEFAULT 'developer',
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			CONSTRAINT valid_role CHECK (role IN ('owner', 'admin', 'developer', 'viewer')),
			CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'pending', 'suspended'))
		)
	`, r.tableName)

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create indexes
	indexes := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_auth_id ON %s (auth_id)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_email ON %s (email)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_username ON %s (username)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_role ON %s (role)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_status ON %s (status)", r.tableName, r.tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_created_at ON %s (created_at)", r.tableName, r.tableName),
	}

	for _, indexQuery := range indexes {
		_, err := r.db.ExecContext(ctx, indexQuery)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// generateUserID generates a unique user ID (simplified implementation)
func generateUserID() string {
	return fmt.Sprintf("usr-%d", time.Now().UnixNano())
}
