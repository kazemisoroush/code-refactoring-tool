// Package services provides business logic for task management
package services

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// TaskService defines the interface for task business logic operations
//
//go:generate mockgen -destination=./mocks/mock_task_service.go -mock_names=TaskService=MockTaskService -package=mocks . TaskService
type TaskService interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.CreateTaskResponse, error)

	// GetTask retrieves a task by its ID
	GetTask(ctx context.Context, taskID string) (*models.GetTaskResponse, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, req *models.UpdateTaskRequest) (*models.UpdateTaskResponse, error)

	// DeleteTask deletes a task by its ID
	DeleteTask(ctx context.Context, taskID string) error

	// ListTasks lists tasks for a project with optional filters
	ListTasks(ctx context.Context, req *models.ListTasksRequest) (*models.ListTasksResponse, error)

	// ExecuteTask executes a task immediately (sync or async)
	ExecuteTask(ctx context.Context, req *models.ExecuteTaskRequest) (*models.ExecuteTaskResponse, error)
}

// NOTE: TaskServiceImpl provides full AI factory integration for task execution
// See task_service_impl.go for the complete implementation
