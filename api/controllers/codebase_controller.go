// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// CodebaseController handles codebase-related HTTP requests
type CodebaseController struct {
	codebaseService services.CodebaseService
}

// NewCodebaseController creates a new CodebaseController
func NewCodebaseController(codebaseService services.CodebaseService) *CodebaseController {
	return &CodebaseController{
		codebaseService: codebaseService,
	}
}

// CreateCodebase handles POST /projects/:project_id/codebases
// @Summary Create a new codebase
// @Description Create a new codebase attached to a project
// @Tags codebases
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body models.CreateCodebaseRequest true "Codebase creation request"
// @Success 201 {object} models.CreateCodebaseResponse "Codebase created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{project_id}/codebases [post]
func (c *CodebaseController) CreateCodebase(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.CreateCodebaseRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to create the codebase
	response, err := c.codebaseService.CreateCodebase(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create codebase",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetCodebase handles GET /codebases/:id
// @Summary Get a codebase by ID
// @Description Retrieve a codebase by its unique identifier
// @Tags codebases
// @Produce json
// @Param id path string true "Codebase ID"
// @Success 200 {object} models.GetCodebaseResponse "Codebase retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid codebase ID"
// @Failure 404 {object} models.ErrorResponse "Codebase not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebases/{id} [get]
func (c *CodebaseController) GetCodebase(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.GetCodebaseRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to get the codebase
	response, err := c.codebaseService.GetCodebase(ctx.Request.Context(), request.CodebaseID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase not found" {
			statusCode = http.StatusNotFound
			message = "Codebase not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to get codebase"
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

// UpdateCodebase handles PUT /codebases/:id
// @Summary Update a codebase
// @Description Update an existing codebase's details
// @Tags codebases
// @Accept json
// @Produce json
// @Param id path string true "Codebase ID"
// @Param request body models.UpdateCodebaseRequest true "Codebase update request"
// @Success 200 {object} models.UpdateCodebaseResponse "Codebase updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Codebase not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebases/{id} [put]
func (c *CodebaseController) UpdateCodebase(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.UpdateCodebaseRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to update the codebase
	response, err := c.codebaseService.UpdateCodebase(ctx.Request.Context(), request)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase not found" {
			statusCode = http.StatusNotFound
			message = "Codebase not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to update codebase"
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

// DeleteCodebase handles DELETE /codebases/:id
// @Summary Delete a codebase
// @Description Delete a codebase by its unique identifier
// @Tags codebases
// @Produce json
// @Param id path string true "Codebase ID"
// @Success 200 {object} models.DeleteCodebaseResponse "Codebase deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid codebase ID"
// @Failure 404 {object} models.ErrorResponse "Codebase not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebases/{id} [delete]
func (c *CodebaseController) DeleteCodebase(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.DeleteCodebaseRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to delete the codebase
	response, err := c.codebaseService.DeleteCodebase(ctx.Request.Context(), request.CodebaseID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "codebase not found" {
			statusCode = http.StatusNotFound
			message = "Codebase not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to delete codebase"
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

// ListCodebases handles GET /codebases
// @Summary List codebases
// @Description Retrieve a list of codebases with optional pagination and filtering
// @Tags codebases
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Param tag_filter query string false "Tag filter in format key:value"
// @Param next_token query string false "Token for pagination"
// @Param max_results query int false "Maximum number of results to return" minimum(1) maximum(100)
// @Success 200 {object} models.ListCodebasesResponse "Codebases retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /codebases [get]
func (c *CodebaseController) ListCodebases(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.ListCodebasesRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to list codebases
	response, err := c.codebaseService.ListCodebases(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list codebases",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
