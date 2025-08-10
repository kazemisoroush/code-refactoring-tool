// Package controllers provides HTTP request handlers for codebase configuration API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// CodebaseConfigController handles codebase configuration-related HTTP requests
type CodebaseConfigController struct {
	codebaseConfigService services.CodebaseConfigService
}

// NewCodebaseConfigController creates a new CodebaseConfigController
func NewCodebaseConfigController(codebaseConfigService services.CodebaseConfigService) *CodebaseConfigController {
	return &CodebaseConfigController{
		codebaseConfigService: codebaseConfigService,
	}
}

// CreateCodebaseConfig handles POST /codebase-configs
// @Summary Create a new codebase configuration
// @Description Create a new codebase configuration profile for reusing across projects
// @Tags codebase-configs
// @Accept json
// @Produce json
// @Param request body models.CreateCodebaseConfigRequest true "Codebase configuration creation request"
// @Success 201 {object} models.CreateCodebaseConfigResponse "Codebase configuration created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebase-configs [post]
func (c *CodebaseConfigController) CreateCodebaseConfig(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.CreateCodebaseConfigRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to create the codebase configuration
	response, err := c.codebaseConfigService.CreateCodebaseConfig(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create codebase configuration",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetCodebaseConfig handles GET /codebase-configs/:config_id
// @Summary Get a codebase configuration by ID
// @Description Retrieve a codebase configuration by its unique identifier
// @Tags codebase-configs
// @Produce json
// @Param config_id path string true "Codebase Configuration ID"
// @Success 200 {object} models.GetCodebaseConfigResponse "Codebase configuration retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid configuration ID"
// @Failure 404 {object} models.ErrorResponse "Codebase configuration not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebase-configs/{config_id} [get]
func (c *CodebaseConfigController) GetCodebaseConfig(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.GetCodebaseConfigRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to get the codebase configuration
	response, err := c.codebaseConfigService.GetCodebaseConfig(ctx.Request.Context(), request.ConfigID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase configuration not found" {
			statusCode = http.StatusNotFound
			message = "Codebase configuration not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to get codebase configuration"
		}

		errorResponse := models.ErrorResponse{
			Code:    statusCode,
			Message: message,
			Details: err.Error(),
		}
		ctx.JSON(statusCode, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateCodebaseConfig handles PUT /codebase-configs/:config_id
// @Summary Update a codebase configuration
// @Description Update an existing codebase configuration's details
// @Tags codebase-configs
// @Accept json
// @Produce json
// @Param config_id path string true "Codebase Configuration ID"
// @Param request body models.UpdateCodebaseConfigRequest true "Codebase configuration update request"
// @Success 200 {object} models.UpdateCodebaseConfigResponse "Codebase configuration updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Codebase configuration not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebase-configs/{config_id} [put]
func (c *CodebaseConfigController) UpdateCodebaseConfig(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.UpdateCodebaseConfigRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to update the codebase configuration
	response, err := c.codebaseConfigService.UpdateCodebaseConfig(ctx.Request.Context(), request)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase configuration not found" {
			statusCode = http.StatusNotFound
			message = "Codebase configuration not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to update codebase configuration"
		}

		errorResponse := models.ErrorResponse{
			Code:    statusCode,
			Message: message,
			Details: err.Error(),
		}
		ctx.JSON(statusCode, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteCodebaseConfig handles DELETE /codebase-configs/:config_id
// @Summary Delete a codebase configuration
// @Description Delete a codebase configuration by its unique identifier
// @Tags codebase-configs
// @Produce json
// @Param config_id path string true "Codebase Configuration ID"
// @Success 200 {object} models.DeleteCodebaseConfigResponse "Codebase configuration deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid configuration ID"
// @Failure 404 {object} models.ErrorResponse "Codebase configuration not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebase-configs/{config_id} [delete]
func (c *CodebaseConfigController) DeleteCodebaseConfig(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.DeleteCodebaseConfigRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to delete the codebase configuration
	response, err := c.codebaseConfigService.DeleteCodebaseConfig(ctx.Request.Context(), request.ConfigID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase configuration not found" {
			statusCode = http.StatusNotFound
			message = "Codebase configuration not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to delete codebase configuration"
		}

		errorResponse := models.ErrorResponse{
			Code:    statusCode,
			Message: message,
			Details: err.Error(),
		}
		ctx.JSON(statusCode, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ListCodebaseConfigs handles GET /codebase-configs
// @Summary List codebase configurations
// @Description Retrieve a list of codebase configurations with optional pagination and filtering
// @Tags codebase-configs
// @Produce json
// @Param next_token query string false "Token for pagination"
// @Param max_results query int false "Maximum number of results to return" minimum(1) maximum(100)
// @Param provider_filter query string false "Filter by provider (github, gitlab, bitbucket, custom)"
// @Param tag_filter query string false "Tag filter in format key:value"
// @Success 200 {object} models.ListCodebaseConfigsResponse "Codebase configurations retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebase-configs [get]
func (c *CodebaseConfigController) ListCodebaseConfigs(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.ListCodebaseConfigsRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to list codebase configurations
	response, err := c.codebaseConfigService.ListCodebaseConfigs(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list codebase configurations",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
