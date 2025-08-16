package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	authService services.AuthService
}

// NewAuthController creates a new auth controller
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// SignUp handles user registration
// @Summary Register a new user
// @Description Register a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.SignUpRequest true "Sign up request"
// @Success 201 {object} models.SignUpResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/signup [post]
func (c *AuthController) SignUp(ctx *gin.Context) {
	var req models.SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := c.authService.SignUp(ctx.Request.Context(), &req)
	if err != nil {
		// Check if it's a user already exists error
		if strings.Contains(err.Error(), "already exists") {
			ctx.JSON(http.StatusConflict, models.ErrorResponse{
				Code:    http.StatusConflict,
				Message: "User already exists",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// SignIn handles user authentication
// @Summary Authenticate a user
// @Description Authenticate a user and return access tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.SignInRequest true "Sign in request"
// @Success 200 {object} models.SignInResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/signin [post]
func (c *AuthController) SignIn(ctx *gin.Context) {
	var req models.SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := c.authService.SignIn(ctx.Request.Context(), &req)
	if err != nil {
		// Check if it's an authentication failure
		if strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "invalid credentials") {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Authentication failed",
				Details: "Invalid username or password",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to sign in",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh an access token using a refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.RefreshTokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := c.authService.RefreshToken(ctx.Request.Context(), &req)
	if err != nil {
		// Check if it's an invalid token error
		if strings.Contains(err.Error(), "invalid token") || strings.Contains(err.Error(), "token refresh failed") {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid refresh token",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to refresh token",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SignOut handles user sign out
// @Summary Sign out a user
// @Description Sign out a user and invalidate tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.SignOutRequest true "Sign out request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/signout [post]
func (c *AuthController) SignOut(ctx *gin.Context) {
	var req models.SignOutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	err := c.authService.SignOut(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to sign out",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Successfully signed out",
	})
}

// CreateUser handles user creation (admin endpoint)
// @Summary Create a new user (Admin)
// @Description Create a new user account (admin operation)
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "Create user request"
// @Success 201 {object} models.CreateUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/users [post]
func (c *AuthController) CreateUser(ctx *gin.Context) {
	var req models.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := c.authService.CreateUser(ctx.Request.Context(), &req)
	if err != nil {
		// Check if it's a user already exists error
		if strings.Contains(err.Error(), "already exists") {
			ctx.JSON(http.StatusConflict, models.ErrorResponse{
				Code:    http.StatusConflict,
				Message: "User already exists",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create user",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetUser handles getting a user by ID
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags authentication
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.GetUserResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/users/{id} [get]
func (c *AuthController) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	response, err := c.authService.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateUser handles updating a user
// @Summary Update user
// @Description Update user information
// @Tags authentication
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "Update user request"
// @Success 200 {object} models.UpdateUserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/users/{id} [put]
func (c *AuthController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	var req models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Set the user ID from the path parameter
	req.UserID = userID

	response, err := c.authService.UpdateUser(ctx.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update user",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteUser handles deleting a user
// @Summary Delete user
// @Description Delete a user account
// @Tags authentication
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/users/{id} [delete]
func (c *AuthController) DeleteUser(ctx *gin.Context) {
	userID := ctx.Param("id")

	err := c.authService.DeleteUser(ctx.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete user",
			Details: err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListUsers handles listing users with optional filtering
// @Summary List users
// @Description List users with optional filtering
// @Tags authentication
// @Produce json
// @Param limit query int false "Maximum number of users to return" default(10)
// @Param offset query int false "Number of users to skip" default(0)
// @Param role query string false "Filter by user role"
// @Param status query string false "Filter by user status"
// @Success 200 {object} models.ListUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/users [get]
func (c *AuthController) ListUsers(ctx *gin.Context) {
	var req models.ListUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	response, err := c.authService.ListUsers(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list users",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ConfirmEmail handles email confirmation after signup
// @Summary Confirm user email
// @Description Verify user account after signup with email verification code
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.ConfirmEmailRequest true "Confirm email request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/confirm [post]
func (c *AuthController) ConfirmEmail(ctx *gin.Context) {
	// Get validated request from context (set by validation middleware)
	validatedRequest, exists := ctx.Get("validatedRequest")
	if !exists {
		// Fallback to direct binding if validation middleware is not used
		var req models.ConfirmEmailRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			})
			return
		}
		validatedRequest = req
	}

	req, ok := validatedRequest.(models.ConfirmEmailRequest)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Details: "Invalid request type",
		})
		return
	}

	err := c.authService.ConfirmEmail(ctx.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to confirm email",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Account verified successfully",
	})
}

// ForgotPassword handles password reset initiation
// @Summary Initiate password reset
// @Description Send password reset code to user's email
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/forgot-password [post]
func (c *AuthController) ForgotPassword(ctx *gin.Context) {
	// Get validated request from context (set by validation middleware)
	validatedRequest, exists := ctx.Get("validatedRequest")
	if !exists {
		// Fallback to direct binding if validation middleware is not used
		var req models.ForgotPasswordRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			})
			return
		}
		validatedRequest = req
	}

	req, ok := validatedRequest.(models.ForgotPasswordRequest)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Details: "Invalid request type",
		})
		return
	}

	err := c.authService.ForgotPassword(ctx.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to initiate password reset",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset code sent to your email",
	})
}

// ResetPassword handles password reset completion
// @Summary Reset password
// @Description Reset user password with confirmation code
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(ctx *gin.Context) {
	// Get validated request from context (set by validation middleware)
	validatedRequest, exists := ctx.Get("validatedRequest")
	if !exists {
		// Fallback to direct binding if validation middleware is not used
		var req models.ResetPasswordRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			})
			return
		}
		validatedRequest = req
	}

	req, ok := validatedRequest.(models.ResetPasswordRequest)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Details: "Invalid request type",
		})
		return
	}

	err := c.authService.ResetPassword(ctx.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reset password",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Password reset successfully",
	})
}

// GetMe handles getting current user information
// @Summary Get current user
// @Description Get the current authenticated user's information
// @Tags authentication
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.APIUser
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /auth/me [get]
func (c *AuthController) GetMe(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "User not authenticated",
			Details: "User ID not found in context",
		})
		return
	}

	// Ensure userID is a string
	userIDStr, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal server error",
			Details: "Invalid user ID type",
		})
		return
	}

	user, err := c.authService.GetUser(ctx.Request.Context(), userIDStr)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "User not found",
				Details: err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get user information",
			Details: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
