package repository

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// TaskRepository defines the interface for task data access operations
type TaskRepository interface {
	// Create creates a new task
	Create(ctx context.Context, task *models.Task) error

	// GetByID retrieves a task by its ID
	GetByID(ctx context.Context, taskID string) (*models.Task, error)

	// Update updates an existing task
	Update(ctx context.Context, task *models.Task) error

	// Delete deletes a task by its ID
	Delete(ctx context.Context, taskID string) error

	// ListByProject lists tasks for a specific project with optional filters
	ListByProject(ctx context.Context, projectID string, filters TaskFilters) ([]models.Task, int, error)

	// ListByAgent lists tasks for a specific agent
	ListByAgent(ctx context.Context, agentID string, filters TaskFilters) ([]models.Task, int, error)

	// ListByCodebase lists tasks for a specific codebase
	ListByCodebase(ctx context.Context, codebaseID string, filters TaskFilters) ([]models.Task, int, error)

	// UpdateStatus updates only the status of a task
	UpdateStatus(ctx context.Context, taskID string, status models.TaskStatus) error

	// UpdateStatusAndOutput updates the status and output of a task
	UpdateStatusAndOutput(ctx context.Context, taskID string, status models.TaskStatus, output map[string]any, errorMessage *string) error
}

// TaskFilters represents filters for task queries
type TaskFilters struct {
	Status     *models.TaskStatus `json:"status,omitempty"`
	Type       *models.TaskType   `json:"type,omitempty"`
	AgentID    *string            `json:"agent_id,omitempty"`
	CodebaseID *string            `json:"codebase_id,omitempty"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}
