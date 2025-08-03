package services

import (
	"context"
	"fmt"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
)

// DefaultTaskService implements TaskService - simplified version for now
type DefaultTaskService struct {
	taskRepo repository.TaskRepository
}

// NewDefaultTaskService creates a new default task service
func NewDefaultTaskService(taskRepo repository.TaskRepository) TaskService {
	return &DefaultTaskService{
		taskRepo: taskRepo,
	}
}

// CreateTask creates a new task
func (s *DefaultTaskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.CreateTaskResponse, error) {
	// TODO: Add validation that project, agent, and codebase exist
	// For now, simplified implementation

	// Create task model
	task := &models.Task{
		ProjectID:   req.ProjectID,
		AgentID:     req.AgentID,
		CodebaseID:  req.CodebaseID,
		Type:        req.Type,
		Status:      models.TaskStatusPending,
		Title:       req.Title,
		Description: req.Description,
		Input:       req.Input,
		Metadata:    req.Metadata,
		Tags:        req.Tags,
	}

	// Save task
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &models.CreateTaskResponse{
		TaskID:    task.TaskID,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}, nil
}

// GetTask retrieves a task by its ID
func (s *DefaultTaskService) GetTask(ctx context.Context, taskID string) (*models.GetTaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &models.GetTaskResponse{
		Task: *task,
	}, nil
}

// UpdateTask updates an existing task
func (s *DefaultTaskService) UpdateTask(ctx context.Context, req *models.UpdateTaskRequest) (*models.UpdateTaskResponse, error) {
	// Get existing task
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update fields if provided
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Output != nil {
		task.Output = req.Output
	}
	if req.ErrorMessage != nil {
		task.ErrorMessage = req.ErrorMessage
	}
	if req.Metadata != nil {
		task.Metadata = req.Metadata
	}

	// Save updated task
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return &models.UpdateTaskResponse{
		Task: *task,
	}, nil
}

// DeleteTask deletes a task by its ID
func (s *DefaultTaskService) DeleteTask(ctx context.Context, taskID string) error {
	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// ListTasks lists tasks for a project with optional filters
func (s *DefaultTaskService) ListTasks(ctx context.Context, req *models.ListTasksRequest) (*models.ListTasksResponse, error) {
	// Set default pagination
	limit := 20
	offset := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	filters := repository.TaskFilters{
		Status:  req.Status,
		Type:    req.Type,
		AgentID: req.AgentID,
		Limit:   limit,
		Offset:  offset,
	}

	tasks, totalCount, err := s.taskRepo.ListByProject(ctx, req.ProjectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return &models.ListTasksResponse{
		Tasks:      tasks,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ExecuteTask executes a task immediately (sync or async)
func (s *DefaultTaskService) ExecuteTask(ctx context.Context, req *models.ExecuteTaskRequest) (*models.ExecuteTaskResponse, error) {
	// First create the task
	createReq := &models.CreateTaskRequest{
		ProjectID:   req.ProjectID,
		AgentID:     req.AgentID,
		CodebaseID:  req.CodebaseID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Input:       req.Input,
	}

	createResp, err := s.CreateTask(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// For now, just return immediately (mock execution)
	// TODO: Implement actual task execution with dynamic AI resources
	output := map[string]any{
		"task_type":    string(req.Type),
		"description":  req.Description,
		"result":       "Task executed successfully (mock)",
		"processed_at": "2024-01-15T10:35:00Z",
	}

	// Update task with mock success
	if err := s.taskRepo.UpdateStatusAndOutput(ctx, createResp.TaskID, models.TaskStatusCompleted, output, nil); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Get updated task
	updatedTask, err := s.taskRepo.GetByID(ctx, createResp.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated task: %w", err)
	}

	return &models.ExecuteTaskResponse{
		TaskID:      createResp.TaskID,
		Status:      updatedTask.Status,
		Output:      updatedTask.Output,
		CreatedAt:   updatedTask.CreatedAt,
		CompletedAt: updatedTask.CompletedAt,
	}, nil
}
