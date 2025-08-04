// Package services provides authentication services that wrap auth providers
package services

import (
	"context"
	"fmt"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
)

//go:generate mockgen -source=auth_service.go -destination=mocks/mock_auth_service.go -package=mocks

// AuthService provides provider-agnostic authentication operations
type AuthService interface {
	// Authentication
	SignUp(ctx context.Context, req *models.SignUpRequest) (*models.SignUpResponse, error)
	SignIn(ctx context.Context, req *models.SignInRequest) (*models.SignInResponse, error)
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.RefreshTokenResponse, error)
	SignOut(ctx context.Context, req *models.SignOutRequest) error

	// Token Operations
	ValidateToken(ctx context.Context, token string) (*models.UserContext, error)

	// User Management (Admin operations)
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.CreateUserResponse, error)
	GetUser(ctx context.Context, userID string) (*models.GetUserResponse, error)
	UpdateUser(ctx context.Context, req *models.UpdateUserRequest) (*models.UpdateUserResponse, error)
	DeleteUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, req *models.ListUsersRequest) (*models.ListUsersResponse, error)
}

// AuthServiceImpl implements AuthService with dependency injection
type AuthServiceImpl struct {
	authProvider auth.AuthProvider
	userRepo     repository.UserRepository
}

// NewAuthService creates a new AuthService with injected dependencies
func NewAuthService(authProvider auth.AuthProvider, userRepo repository.UserRepository) AuthService {
	return &AuthServiceImpl{
		authProvider: authProvider,
		userRepo:     userRepo,
	}
}

// SignIn authenticates a user and returns tokens
func (s *AuthServiceImpl) SignIn(ctx context.Context, req *models.SignInRequest) (*models.SignInResponse, error) {
	// Use the abstract auth provider (could be Cognito, Auth0, Firebase, etc.)
	authReq := &auth.SignInRequest{
		Username: req.Username,
		Password: req.Password,
	}

	authResult, err := s.authProvider.SignIn(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Sync user to database (create if not exists, update if exists)
	dbUser, err := s.syncUserToDatabase(ctx, authResult.User)
	if err != nil {
		return nil, fmt.Errorf("failed to sync user to database: %w", err)
	}

	return &models.SignInResponse{
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		TokenType:    authResult.TokenType,
		ExpiresIn:    authResult.ExpiresIn,
		User:         s.mapToAPIUser(dbUser),
	}, nil
}

// ValidateToken validates a token and returns user context
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, token string) (*models.UserContext, error) {
	// Provider-agnostic token validation
	claims, err := s.authProvider.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Get user from our database using the auth provider's user ID
	user, err := s.userRepo.GetUserByAuthID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found in database: %w", err)
	}

	return &models.UserContext{
		UserID:      user.UserID,
		AuthID:      claims.UserID,
		Email:       user.Email,
		Username:    user.Username,
		Role:        user.Role,
		Permissions: s.getUserPermissions(user.Role),
	}, nil
}

// SignUp creates a new user account
func (s *AuthServiceImpl) SignUp(ctx context.Context, req *models.SignUpRequest) (*models.SignUpResponse, error) {
	// Create user in auth provider
	authReq := &auth.SignUpRequest{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	authResult, err := s.authProvider.SignUp(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("user creation failed: %w", err)
	}

	// Create user in our database
	dbUser, err := s.syncUserToDatabase(ctx, authResult.User)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in database: %w", err)
	}

	return &models.SignUpResponse{
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		TokenType:    authResult.TokenType,
		ExpiresIn:    authResult.ExpiresIn,
		User:         s.mapToAPIUser(dbUser),
	}, nil
}

// RefreshToken refreshes an access token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.RefreshTokenResponse, error) {
	authResult, err := s.authProvider.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	return &models.RefreshTokenResponse{
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		TokenType:    authResult.TokenType,
		ExpiresIn:    authResult.ExpiresIn,
	}, nil
}

// SignOut signs out a user
func (s *AuthServiceImpl) SignOut(ctx context.Context, req *models.SignOutRequest) error {
	err := s.authProvider.SignOut(ctx, req.AccessToken)
	if err != nil {
		return fmt.Errorf("sign out failed: %w", err)
	}
	return nil
}

// CreateUser creates a new user (admin operation)
func (s *AuthServiceImpl) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.CreateUserResponse, error) {
	// Create in auth provider first
	authReq := &auth.CreateUserRequest{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	authUser, err := s.authProvider.CreateUser(ctx, authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in auth provider: %w", err)
	}

	// Create in our database
	dbUser, err := s.syncUserToDatabase(ctx, authUser)
	if err != nil {
		return nil, fmt.Errorf("failed to sync user to database: %w", err)
	}

	return &models.CreateUserResponse{
		User: s.mapToAPIUser(dbUser),
	}, nil
}

// GetUser retrieves a user by ID
func (s *AuthServiceImpl) GetUser(ctx context.Context, userID string) (*models.GetUserResponse, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &models.GetUserResponse{
		User: s.mapToAPIUser(user),
	}, nil
}

// UpdateUser updates a user
func (s *AuthServiceImpl) UpdateUser(ctx context.Context, req *models.UpdateUserRequest) (*models.UpdateUserResponse, error) {
	// Get current user from database
	user, err := s.userRepo.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update in auth provider
	authReq := &auth.UpdateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	authUser, err := s.authProvider.UpdateUser(ctx, user.AuthID, authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update user in auth provider: %w", err)
	}

	// Update in our database
	updatedUser, err := s.syncUserToDatabase(ctx, authUser)
	if err != nil {
		return nil, fmt.Errorf("failed to sync user to database: %w", err)
	}

	return &models.UpdateUserResponse{
		User: s.mapToAPIUser(updatedUser),
	}, nil
}

// DeleteUser deletes a user
func (s *AuthServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	// Get user from database
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Delete from auth provider
	err = s.authProvider.DeleteUser(ctx, user.AuthID)
	if err != nil {
		return fmt.Errorf("failed to delete user from auth provider: %w", err)
	}

	// Delete from our database
	err = s.userRepo.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user from database: %w", err)
	}

	return nil
}

// ListUsers lists users
func (s *AuthServiceImpl) ListUsers(ctx context.Context, req *models.ListUsersRequest) (*models.ListUsersResponse, error) {
	users, total, err := s.userRepo.ListUsers(ctx, &repository.ListUsersFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
		Role:   req.Role,
		Status: req.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	apiUsers := make([]*models.APIUser, len(users))
	for i, user := range users {
		apiUsers[i] = s.mapToAPIUser(user)
	}

	return &models.ListUsersResponse{
		Users:  apiUsers,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// Private helper methods

// syncUserToDatabase creates or updates a user in our database based on auth provider user
func (s *AuthServiceImpl) syncUserToDatabase(ctx context.Context, authUser *auth.User) (*models.DBUser, error) {
	// Try to get existing user
	existingUser, err := s.userRepo.GetUserByAuthID(ctx, authUser.ID)
	if err != nil {
		// User doesn't exist, create new one
		newUser := &models.DBUser{
			AuthID:    authUser.ID,
			Email:     authUser.Email,
			Username:  authUser.Username,
			FirstName: authUser.FirstName,
			LastName:  authUser.LastName,
			Role:      models.RoleDeveloper, // Default role
			Status:    s.mapAuthStatusToDBStatus(authUser.Status),
		}

		return s.userRepo.CreateUser(ctx, newUser)
	}

	// User exists, update if needed
	existingUser.Email = authUser.Email
	existingUser.Username = authUser.Username
	existingUser.FirstName = authUser.FirstName
	existingUser.LastName = authUser.LastName
	existingUser.Status = s.mapAuthStatusToDBStatus(authUser.Status)

	return s.userRepo.UpdateUser(ctx, existingUser)
}

// mapToAPIUser converts database user to API user
func (s *AuthServiceImpl) mapToAPIUser(dbUser *models.DBUser) *models.APIUser {
	return &models.APIUser{
		UserID:    dbUser.UserID,
		Email:     dbUser.Email,
		Username:  dbUser.Username,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		Role:      dbUser.Role,
		Status:    dbUser.Status,
		CreatedAt: dbUser.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: dbUser.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// mapAuthStatusToDBStatus converts auth provider status to database status
func (s *AuthServiceImpl) mapAuthStatusToDBStatus(authStatus auth.UserStatus) models.UserStatus {
	switch authStatus {
	case auth.UserStatusActive:
		return models.UserStatusActive
	case auth.UserStatusInactive:
		return models.UserStatusInactive
	case auth.UserStatusPending:
		return models.UserStatusPending
	case auth.UserStatusSuspended:
		return models.UserStatusSuspended
	default:
		return models.UserStatusInactive
	}
}

// getUserPermissions returns permissions based on user role
func (s *AuthServiceImpl) getUserPermissions(role models.UserRole) []models.Permission {
	switch role {
	case models.RoleOwner:
		return []models.Permission{
			models.PermissionProjectCreate, models.PermissionProjectRead,
			models.PermissionProjectUpdate, models.PermissionProjectDelete,
			models.PermissionTaskCreate, models.PermissionTaskExecute,
			models.PermissionTaskRead, models.PermissionTaskUpdate, models.PermissionTaskDelete,
			models.PermissionAgentCreate, models.PermissionAgentRead,
			models.PermissionAgentUpdate, models.PermissionAgentDelete,
			models.PermissionUserCreate, models.PermissionUserRead,
			models.PermissionUserUpdate, models.PermissionUserDelete,
		}
	case models.RoleAdmin:
		return []models.Permission{
			models.PermissionProjectCreate, models.PermissionProjectRead,
			models.PermissionProjectUpdate, models.PermissionProjectDelete,
			models.PermissionTaskCreate, models.PermissionTaskExecute,
			models.PermissionTaskRead, models.PermissionTaskUpdate, models.PermissionTaskDelete,
			models.PermissionAgentCreate, models.PermissionAgentRead,
			models.PermissionAgentUpdate, models.PermissionAgentDelete,
			models.PermissionUserRead, models.PermissionUserUpdate,
		}
	case models.RoleDeveloper:
		return []models.Permission{
			models.PermissionProjectCreate, models.PermissionProjectRead,
			models.PermissionProjectUpdate,
			models.PermissionTaskCreate, models.PermissionTaskExecute, models.PermissionTaskRead,
			models.PermissionAgentCreate, models.PermissionAgentRead,
		}
	case models.RoleViewer:
		return []models.Permission{
			models.PermissionProjectRead,
			models.PermissionTaskRead,
			models.PermissionAgentRead,
		}
	default:
		return []models.Permission{}
	}
}
