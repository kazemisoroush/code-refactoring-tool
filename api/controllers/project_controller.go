// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// ProjectController handles project-related HTTP requests
type ProjectController struct {
	projectService services.ProjectService
}

// NewProjectController creates a new ProjectController
func NewProjectController(projectService services.ProjectService) *ProjectController {
	return &ProjectController{
		projectService: projectService,
	}
}

// CreateProject handles POST /projects
// @Summary Create a new project
// @Description Create a new project for organizing codebases and agents
// @Tags projects
// @Accept json
// @Produce json
// @Param request body models.CreateProjectRequest true "Project creation request"
// @Success 201 {object} models.CreateProjectResponse "Project created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects [post]
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.CreateProjectRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to create the project
	response, err := c.projectService.CreateProject(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create project",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetProject handles GET /projects/:id
// @Summary Get a project by ID
// @Description Retrieve a project by its unique identifier
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.GetProjectResponse "Project retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid project ID"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [get]
func (c *ProjectController) GetProject(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.GetProjectRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to get the project
	response, err := c.projectService.GetProject(ctx.Request.Context(), request.ProjectID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "project not found" {
			statusCode = http.StatusNotFound
			message = "Project not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to get project"
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

// UpdateProject handles PUT /projects/:id
// @Summary Update a project
// @Description Update an existing project's details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body models.UpdateProjectRequest true "Project update request"
// @Success 200 {object} models.UpdateProjectResponse "Project updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [put]
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.UpdateProjectRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to update the project
	response, err := c.projectService.UpdateProject(ctx.Request.Context(), request)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "project not found" {
			statusCode = http.StatusNotFound
			message = "Project not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to update project"
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

// DeleteProject handles DELETE /projects/:id
// @Summary Delete a project
// @Description Delete a project by its unique identifier
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.DeleteProjectResponse "Project deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid project ID"
// @Failure 404 {object} models.ErrorResponse "Project not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects/{id} [delete]
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.DeleteProjectRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to delete the project
	response, err := c.projectService.DeleteProject(ctx.Request.Context(), request.ProjectID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "project not found" {
			statusCode = http.StatusNotFound
			message = "Project not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to delete project"
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

// ListProjects handles GET /projects
// @Summary List projects
// @Description Retrieve a list of projects with optional pagination and filtering
// @Tags projects
// @Produce json
// @Param next_token query string false "Token for pagination"
// @Param max_results query int false "Maximum number of results to return" minimum(1) maximum(100)
// @Param tag_filter query string false "Tag filter in format key:value"
// @Success 200 {object} models.ListProjectsResponse "Projects retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /projects [get]
func (c *ProjectController) ListProjects(ctx *gin.Context) {
	// Get the validated request from context (set by validation middleware)
	request, exists := middleware.GetValidatedRequest[models.ListProjectsRequest](ctx)
	if !exists {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Missing validated request",
			Details: "Validation middleware must be applied before this controller",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to list projects
	response, err := c.projectService.ListProjects(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list projects",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
