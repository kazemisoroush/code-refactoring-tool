package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository/mocks"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
	authMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/auth/mocks"
)

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestNewAuthService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	service := NewAuthService(mockAuthProvider, mockUserRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &AuthServiceImpl{}, service)
}

func TestAuthServiceImpl_SignIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful sign in with existing user", func(t *testing.T) {
		req := &models.SignInRequest{
			Username: "testuser",
			Password: "password123",
		}

		authUser := &auth.User{
			ID:        "auth-123",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Status:    auth.UserStatusActive,
		}

		authResult := &auth.AuthResult{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User:         authUser,
		}

		existingUser := &models.DBUser{
			UserID:    "user-123",
			AuthID:    "auth-123",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockAuthProvider.EXPECT().
			SignIn(ctx, &auth.SignInRequest{
				Username: "testuser",
				Password: "password123",
			}).
			Return(authResult, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-123").
			Return(existingUser, nil)

		mockUserRepo.EXPECT().
			UpdateUser(ctx, gomock.Any()).
			Return(existingUser, nil)

		result, err := service.SignIn(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "access-token", result.AccessToken)
		assert.Equal(t, "refresh-token", result.RefreshToken)
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Equal(t, 3600, result.ExpiresIn)
		assert.Equal(t, "user-123", result.User.UserID)
		assert.Equal(t, "testuser", result.User.Username)
		assert.Equal(t, "test@example.com", result.User.Email)
	})

	t.Run("successful sign in with new user", func(t *testing.T) {
		req := &models.SignInRequest{
			Username: "newuser",
			Password: "password123",
		}

		authUser := &auth.User{
			ID:        "auth-456",
			Username:  "newuser",
			Email:     "new@example.com",
			FirstName: stringPtr("New"),
			LastName:  stringPtr("User"),
			Status:    auth.UserStatusActive,
		}

		authResult := &auth.AuthResult{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User:         authUser,
		}

		newUser := &models.DBUser{
			UserID:    "user-456",
			AuthID:    "auth-456",
			Username:  "newuser",
			Email:     "new@example.com",
			FirstName: stringPtr("New"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockAuthProvider.EXPECT().
			SignIn(ctx, &auth.SignInRequest{
				Username: "newuser",
				Password: "password123",
			}).
			Return(authResult, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-456").
			Return(nil, errors.New("user not found"))

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Do(func(_ context.Context, user *models.DBUser) {
				assert.Equal(t, "auth-456", user.AuthID)
				assert.Equal(t, "newuser", user.Username)
				assert.Equal(t, "new@example.com", user.Email)
				assert.Equal(t, models.RoleDeveloper, user.Role)
				assert.Equal(t, models.UserStatusActive, user.Status)
			}).
			Return(newUser, nil)

		result, err := service.SignIn(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "access-token", result.AccessToken)
		assert.Equal(t, "user-456", result.User.UserID)
	})

	t.Run("auth provider error", func(t *testing.T) {
		req := &models.SignInRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}

		mockAuthProvider.EXPECT().
			SignIn(ctx, gomock.Any()).
			Return(nil, errors.New("invalid credentials"))

		result, err := service.SignIn(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("database sync error", func(t *testing.T) {
		req := &models.SignInRequest{
			Username: "testuser",
			Password: "password123",
		}

		authUser := &auth.User{
			ID:       "auth-123",
			Username: "testuser",
			Email:    "test@example.com",
			Status:   auth.UserStatusActive,
		}

		authResult := &auth.AuthResult{
			AccessToken: "access-token",
			User:        authUser,
		}

		mockAuthProvider.EXPECT().
			SignIn(ctx, gomock.Any()).
			Return(authResult, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-123").
			Return(nil, errors.New("user not found"))

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil, errors.New("database error"))

		result, err := service.SignIn(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to sync user to database")
	})
}

func TestAuthServiceImpl_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful token validation", func(t *testing.T) {
		token := "valid-token"

		claims := &auth.TokenClaims{
			UserID:   "auth-123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		user := &models.DBUser{
			UserID:    "user-123",
			AuthID:    "auth-123",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Role:      models.RoleAdmin,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockAuthProvider.EXPECT().
			ValidateToken(ctx, token).
			Return(claims, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-123").
			Return(user, nil)

		result, err := service.ValidateToken(ctx, token)

		require.NoError(t, err)
		assert.Equal(t, "user-123", result.UserID)
		assert.Equal(t, "auth-123", result.AuthID)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Equal(t, "testuser", result.Username)
		assert.Equal(t, models.RoleAdmin, result.Role)
		assert.NotEmpty(t, result.Permissions)
	})

	t.Run("invalid token", func(t *testing.T) {
		token := "invalid-token"

		mockAuthProvider.EXPECT().
			ValidateToken(ctx, token).
			Return(nil, errors.New("invalid token"))

		result, err := service.ValidateToken(ctx, token)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token validation failed")
	})

	t.Run("user not found in database", func(t *testing.T) {
		token := "valid-token"

		claims := &auth.TokenClaims{
			UserID: "auth-123",
		}

		mockAuthProvider.EXPECT().
			ValidateToken(ctx, token).
			Return(claims, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-123").
			Return(nil, errors.New("user not found"))

		result, err := service.ValidateToken(ctx, token)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found in database")
	})
}

func TestAuthServiceImpl_SignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful sign up", func(t *testing.T) {
		req := &models.SignUpRequest{
			Username:  "newuser",
			Email:     "new@example.com",
			Password:  "password123",
			FirstName: stringPtr("New"),
			LastName:  stringPtr("User"),
		}

		authUser := &auth.User{
			ID:        "auth-456",
			Username:  "newuser",
			Email:     "new@example.com",
			FirstName: stringPtr("New"),
			LastName:  stringPtr("User"),
			Status:    auth.UserStatusActive,
		}

		authResult := &auth.AuthResult{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User:         authUser,
		}

		dbUser := &models.DBUser{
			UserID:    "user-456",
			AuthID:    "auth-456",
			Username:  "newuser",
			Email:     "new@example.com",
			FirstName: stringPtr("New"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockAuthProvider.EXPECT().
			SignUp(ctx, &auth.SignUpRequest{
				Username:  "newuser",
				Email:     "new@example.com",
				Password:  "password123",
				FirstName: stringPtr("New"),
				LastName:  stringPtr("User"),
			}).
			Return(authResult, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-456").
			Return(nil, errors.New("user not found"))

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(dbUser, nil)

		result, err := service.SignUp(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "access-token", result.AccessToken)
		assert.Equal(t, "refresh-token", result.RefreshToken)
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Equal(t, 3600, result.ExpiresIn)
		assert.Equal(t, "user-456", result.User.UserID)
	})

	t.Run("auth provider error", func(t *testing.T) {
		req := &models.SignUpRequest{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "password123",
		}

		mockAuthProvider.EXPECT().
			SignUp(ctx, gomock.Any()).
			Return(nil, errors.New("email already exists"))

		result, err := service.SignUp(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user creation failed")
	})
}

func TestAuthServiceImpl_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()

	t.Run("successful token refresh", func(t *testing.T) {
		req := &models.RefreshTokenRequest{
			RefreshToken: "refresh-token",
		}

		authResult := &auth.AuthResult{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
		}

		mockAuthProvider.EXPECT().
			RefreshToken(ctx, "refresh-token").
			Return(authResult, nil)

		result, err := service.RefreshToken(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "new-access-token", result.AccessToken)
		assert.Equal(t, "new-refresh-token", result.RefreshToken)
		assert.Equal(t, "Bearer", result.TokenType)
		assert.Equal(t, 3600, result.ExpiresIn)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		req := &models.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}

		mockAuthProvider.EXPECT().
			RefreshToken(ctx, "invalid-token").
			Return(nil, errors.New("invalid refresh token"))

		result, err := service.RefreshToken(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token refresh failed")
	})
}

func TestAuthServiceImpl_SignOut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()

	t.Run("successful sign out", func(t *testing.T) {
		req := &models.SignOutRequest{
			AccessToken: "access-token",
		}

		mockAuthProvider.EXPECT().
			SignOut(ctx, "access-token").
			Return(nil)

		err := service.SignOut(ctx, req)

		assert.NoError(t, err)
	})

	t.Run("sign out error", func(t *testing.T) {
		req := &models.SignOutRequest{
			AccessToken: "invalid-token",
		}

		mockAuthProvider.EXPECT().
			SignOut(ctx, "invalid-token").
			Return(errors.New("invalid token"))

		err := service.SignOut(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sign out failed")
	})
}

func TestAuthServiceImpl_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful user creation", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Username:  "adminuser",
			Email:     "admin@example.com",
			FirstName: stringPtr("Admin"),
			LastName:  stringPtr("User"),
		}

		authUser := &auth.User{
			ID:        "auth-789",
			Username:  "adminuser",
			Email:     "admin@example.com",
			FirstName: stringPtr("Admin"),
			LastName:  stringPtr("User"),
			Status:    auth.UserStatusActive,
		}

		dbUser := &models.DBUser{
			UserID:    "user-789",
			AuthID:    "auth-789",
			Username:  "adminuser",
			Email:     "admin@example.com",
			FirstName: stringPtr("Admin"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockAuthProvider.EXPECT().
			CreateUser(ctx, &auth.CreateUserRequest{
				Username:  "adminuser",
				Email:     "admin@example.com",
				FirstName: stringPtr("Admin"),
				LastName:  stringPtr("User"),
			}).
			Return(authUser, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-789").
			Return(nil, errors.New("user not found"))

		mockUserRepo.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(dbUser, nil)

		result, err := service.CreateUser(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "user-789", result.User.UserID)
		assert.Equal(t, "adminuser", result.User.Username)
		assert.Equal(t, "admin@example.com", result.User.Email)
	})

	t.Run("auth provider error", func(t *testing.T) {
		req := &models.CreateUserRequest{
			Username: "adminuser",
			Email:    "admin@example.com",
		}

		mockAuthProvider.EXPECT().
			CreateUser(ctx, gomock.Any()).
			Return(nil, errors.New("user already exists"))

		result, err := service.CreateUser(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user in auth provider")
	})
}

func TestAuthServiceImpl_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful get user", func(t *testing.T) {
		userID := "user-123"

		user := &models.DBUser{
			UserID:    "user-123",
			AuthID:    "auth-123",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(user, nil)

		result, err := service.GetUser(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, "user-123", result.User.UserID)
		assert.Equal(t, "testuser", result.User.Username)
		assert.Equal(t, "test@example.com", result.User.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent"

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(nil, errors.New("user not found"))

		result, err := service.GetUser(ctx, userID)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
	})
}

func TestAuthServiceImpl_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful user update", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			UserID:    "user-123",
			Email:     stringPtr("updated@example.com"),
			FirstName: stringPtr("Updated"),
			LastName:  stringPtr("User"),
		}

		existingUser := &models.DBUser{
			UserID:    "user-123",
			AuthID:    "auth-123",
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		authUser := &auth.User{
			ID:        "auth-123",
			Username:  "testuser",
			Email:     "updated@example.com",
			FirstName: stringPtr("Updated"),
			LastName:  stringPtr("User"),
			Status:    auth.UserStatusActive,
		}

		updatedUser := &models.DBUser{
			UserID:    "user-123",
			AuthID:    "auth-123",
			Username:  "testuser",
			Email:     "updated@example.com",
			FirstName: stringPtr("Updated"),
			LastName:  stringPtr("User"),
			Role:      models.RoleDeveloper,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, "user-123").
			Return(existingUser, nil)

		mockAuthProvider.EXPECT().
			UpdateUser(ctx, "auth-123", &auth.UpdateUserRequest{
				Email:     stringPtr("updated@example.com"),
				FirstName: stringPtr("Updated"),
				LastName:  stringPtr("User"),
			}).
			Return(authUser, nil)

		mockUserRepo.EXPECT().
			GetUserByAuthID(ctx, "auth-123").
			Return(existingUser, nil)

		mockUserRepo.EXPECT().
			UpdateUser(ctx, gomock.Any()).
			Return(updatedUser, nil)

		result, err := service.UpdateUser(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "user-123", result.User.UserID)
		assert.Equal(t, "updated@example.com", result.User.Email)
		assert.Equal(t, "Updated", *result.User.FirstName)
	})

	t.Run("user not found", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			UserID: "nonexistent",
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, "nonexistent").
			Return(nil, errors.New("user not found"))

		result, err := service.UpdateUser(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
	})

	t.Run("auth provider update error", func(t *testing.T) {
		req := &models.UpdateUserRequest{
			UserID: "user-123",
			Email:  stringPtr("updated@example.com"),
		}

		existingUser := &models.DBUser{
			UserID: "user-123",
			AuthID: "auth-123",
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, "user-123").
			Return(existingUser, nil)

		mockAuthProvider.EXPECT().
			UpdateUser(ctx, "auth-123", gomock.Any()).
			Return(nil, errors.New("update failed"))

		result, err := service.UpdateUser(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user in auth provider")
	})
}

func TestAuthServiceImpl_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()

	t.Run("successful user deletion", func(t *testing.T) {
		userID := "user-123"

		user := &models.DBUser{
			UserID: "user-123",
			AuthID: "auth-123",
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(user, nil)

		mockAuthProvider.EXPECT().
			DeleteUser(ctx, "auth-123").
			Return(nil)

		mockUserRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil)

		err := service.DeleteUser(ctx, userID)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		userID := "nonexistent"

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(nil, errors.New("user not found"))

		err := service.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get user")
	})

	t.Run("auth provider delete error", func(t *testing.T) {
		userID := "user-123"

		user := &models.DBUser{
			UserID: "user-123",
			AuthID: "auth-123",
		}

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(user, nil)

		mockAuthProvider.EXPECT().
			DeleteUser(ctx, "auth-123").
			Return(errors.New("delete failed"))

		err := service.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete user from auth provider")
	})
}

func TestAuthServiceImpl_ListUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := NewAuthService(mockAuthProvider, mockUserRepo)

	ctx := context.Background()
	now := time.Now()

	t.Run("successful list users", func(t *testing.T) {
		roleDeveloper := models.RoleDeveloper
		req := &models.ListUsersRequest{
			Limit:  10,
			Offset: 0,
			Role:   &roleDeveloper,
		}

		users := []*models.DBUser{
			{
				UserID:    "user-1",
				Username:  "user1",
				Email:     "user1@example.com",
				FirstName: stringPtr("User"),
				LastName:  stringPtr("One"),
				Role:      models.RoleDeveloper,
				Status:    models.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				UserID:    "user-2",
				Username:  "user2",
				Email:     "user2@example.com",
				FirstName: stringPtr("User"),
				LastName:  stringPtr("Two"),
				Role:      models.RoleDeveloper,
				Status:    models.UserStatusActive,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		mockUserRepo.EXPECT().
			ListUsers(ctx, &repository.ListUsersFilter{
				Limit:  10,
				Offset: 0,
				Role:   &roleDeveloper,
			}).
			Return(users, 2, nil)

		result, err := service.ListUsers(ctx, req)

		require.NoError(t, err)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 0, result.Offset)
		assert.Equal(t, "user-1", result.Users[0].UserID)
		assert.Equal(t, "user-2", result.Users[1].UserID)
	})

	t.Run("repository error", func(t *testing.T) {
		req := &models.ListUsersRequest{
			Limit:  10,
			Offset: 0,
		}

		mockUserRepo.EXPECT().
			ListUsers(ctx, gomock.Any()).
			Return(nil, 0, errors.New("database error"))

		result, err := service.ListUsers(ctx, req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list users")
	})
}

func TestAuthServiceImpl_mapAuthStatusToDBStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := &AuthServiceImpl{
		authProvider: mockAuthProvider,
		userRepo:     mockUserRepo,
	}

	tests := []struct {
		name       string
		authStatus auth.UserStatus
		expected   models.UserStatus
	}{
		{"Active", auth.UserStatusActive, models.UserStatusActive},
		{"Inactive", auth.UserStatusInactive, models.UserStatusInactive},
		{"Pending", auth.UserStatusPending, models.UserStatusPending},
		{"Suspended", auth.UserStatusSuspended, models.UserStatusSuspended},
		{"Unknown", auth.UserStatus("unknown"), models.UserStatusInactive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.mapAuthStatusToDBStatus(tt.authStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthServiceImpl_getUserPermissions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := &AuthServiceImpl{
		authProvider: mockAuthProvider,
		userRepo:     mockUserRepo,
	}

	tests := []struct {
		name     string
		role     models.UserRole
		expected int // Number of expected permissions
	}{
		{"Owner", models.RoleOwner, 17},            // All permissions
		{"Admin", models.RoleAdmin, 15},            // Most permissions except user create/delete
		{"Developer", models.RoleDeveloper, 8},     // Basic development permissions
		{"Viewer", models.RoleViewer, 3},           // Read-only permissions
		{"Unknown", models.UserRole("unknown"), 0}, // No permissions
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := service.getUserPermissions(tt.role)
			assert.Len(t, permissions, tt.expected)

			// Verify specific permissions for each role
			switch tt.role {
			case models.RoleOwner:
				assert.Contains(t, permissions, models.PermissionUserCreate)
				assert.Contains(t, permissions, models.PermissionUserDelete)
				assert.Contains(t, permissions, models.PermissionProjectDelete)
			case models.RoleAdmin:
				assert.NotContains(t, permissions, models.PermissionUserCreate)
				assert.NotContains(t, permissions, models.PermissionUserDelete)
				assert.Contains(t, permissions, models.PermissionProjectDelete)
			case models.RoleDeveloper:
				assert.Contains(t, permissions, models.PermissionProjectCreate)
				assert.NotContains(t, permissions, models.PermissionProjectDelete)
				assert.NotContains(t, permissions, models.PermissionUserCreate)
			case models.RoleViewer:
				assert.Contains(t, permissions, models.PermissionProjectRead)
				assert.NotContains(t, permissions, models.PermissionProjectCreate)
			}
		})
	}
}

func TestAuthServiceImpl_mapToAPIUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthProvider := authMocks.NewMockAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	service := &AuthServiceImpl{
		authProvider: mockAuthProvider,
		userRepo:     mockUserRepo,
	}

	now := time.Now()
	dbUser := &models.DBUser{
		UserID:    "user-123",
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: stringPtr("Test"),
		LastName:  stringPtr("User"),
		Role:      models.RoleDeveloper,
		Status:    models.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	apiUser := service.mapToAPIUser(dbUser)

	assert.Equal(t, "user-123", apiUser.UserID)
	assert.Equal(t, "test@example.com", apiUser.Email)
	assert.Equal(t, "testuser", apiUser.Username)
	assert.Equal(t, "Test", *apiUser.FirstName)
	assert.Equal(t, "User", *apiUser.LastName)
	assert.Equal(t, models.RoleDeveloper, apiUser.Role)
	assert.Equal(t, models.UserStatusActive, apiUser.Status)
	assert.Equal(t, now.Format("2006-01-02T15:04:05Z"), apiUser.CreatedAt)
	assert.Equal(t, now.Format("2006-01-02T15:04:05Z"), apiUser.UpdatedAt)
}
