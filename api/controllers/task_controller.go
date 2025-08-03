package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// TaskController handles HTTP requests for task operations
type TaskController struct {
	taskService services.TaskService
}

// NewTaskController creates a new task controller
func NewTaskController(taskService services.TaskService) *TaskController {
	return &TaskController{
		taskService: taskService,
	}
}

// CreateTask creates a new task
// @Summary Create a new task
// @Description Create a new task for a specific project and agent
// @Tags tasks
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body models.CreateTaskRequest true "Task creation request"
// @Success 201 {object} models.CreateTaskResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{project_id}/tasks [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	var req models.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project ID from URL path
	projectID := ctx.Param("project_id")
	req.ProjectID = projectID

	response, err := c.taskService.CreateTask(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetTask retrieves a task by ID
// @Summary Get a task by ID
// @Description Retrieve a task by its unique identifier
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.GetTaskResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id} [get]
func (c *TaskController) GetTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	response, err := c.taskService.GetTask(ctx.Request.Context(), taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateTask updates an existing task
// @Summary Update a task
// @Description Update an existing task's status, output, or metadata
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.UpdateTaskRequest true "Task update request"
// @Success 200 {object} models.UpdateTaskResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id} [put]
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	var req models.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get task ID from URL path
	req.TaskID = ctx.Param("id")

	response, err := c.taskService.UpdateTask(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteTask deletes a task by ID
// @Summary Delete a task
// @Description Delete a task by its unique identifier
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tasks/{id} [delete]
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	err := c.taskService.DeleteTask(ctx.Request.Context(), taskID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListTasks lists tasks for a project
// @Summary List tasks for a project
// @Description List all tasks for a specific project with optional filtering
// @Tags tasks
// @Produce json
// @Param project_id path string true "Project ID"
// @Param status query string false "Filter by task status" Enums(pending, in_progress, completed, failed, cancelled)
// @Param type query string false "Filter by task type" Enums(code_analysis, refactoring, code_review, documentation, custom)
// @Param agent_id query string false "Filter by agent ID"
// @Param limit query int false "Number of results to return (default 20, max 100)"
// @Param offset query int false "Number of results to skip (default 0)"
// @Success 200 {object} models.ListTasksResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{project_id}/tasks [get]
func (c *TaskController) ListTasks(ctx *gin.Context) {
	var req models.ListTasksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project ID from URL path
	req.ProjectID = ctx.Param("project_id")

	response, err := c.taskService.ListTasks(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ExecuteTask executes a task immediately
// @Summary Execute a task immediately
// @Description Create and execute a task immediately (synchronously or asynchronously)
// @Tags tasks
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body models.ExecuteTaskRequest true "Task execution request"
// @Success 200 {object} models.ExecuteTaskResponse "Synchronous execution completed"
// @Success 202 {object} models.ExecuteTaskResponse "Asynchronous execution started"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/projects/{project_id}/tasks/execute [post]
func (c *TaskController) ExecuteTask(ctx *gin.Context) {
	var req models.ExecuteTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project ID from URL path
	req.ProjectID = ctx.Param("project_id")

	response, err := c.taskService.ExecuteTask(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return 202 for async execution, 200 for sync
	statusCode := http.StatusOK
	if req.Async {
		statusCode = http.StatusAccepted
	}

	ctx.JSON(statusCode, response)
}
