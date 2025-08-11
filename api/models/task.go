// Package models provides data structures for task execution and management
package models

import (
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	// TaskStatusPending indicates the task is waiting to be processed
	TaskStatusPending TaskStatus = "pending"

	// TaskStatusInProgress indicates the task is currently being executed
	TaskStatusInProgress TaskStatus = "in_progress"

	// TaskStatusCompleted indicates the task has been successfully completed
	TaskStatusCompleted TaskStatus = "completed"

	// TaskStatusFailed indicates the task has failed
	TaskStatusFailed TaskStatus = "failed"

	// TaskStatusCancelled indicates the task has been cancelled by the user
	TaskStatusCancelled TaskStatus = "cancelled"
)

// TaskType represents the type of task to execute
type TaskType string

const (
	// TaskTypeCodeAnalysis represents a code analysis task
	TaskTypeCodeAnalysis TaskType = "code_analysis"

	// TaskTypeRefactoring represents a code refactoring task
	TaskTypeRefactoring TaskType = "refactoring"

	// TaskTypeCodeReview represents a code review task
	TaskTypeCodeReview TaskType = "code_review"

	// TaskTypeDocumentation represents a documentation task
	TaskTypeDocumentation TaskType = "documentation"

	// TaskTypeCustom represents a user-defined custom task
	TaskTypeCustom TaskType = "custom"
)

// Task represents a user-initiated task/prompt execution against a project
type Task struct {
	TaskID       string            `json:"task_id" db:"task_id"`
	ProjectID    string            `json:"project_id" db:"project_id"`
	AgentID      string            `json:"agent_id" db:"agent_id"`                 // Which agent to use for this task
	CodebaseID   *string           `json:"codebase_id,omitempty" db:"codebase_id"` // Optional: specific codebase, if nil uses all project codebases
	Type         TaskType          `json:"type" db:"type"`
	Status       TaskStatus        `json:"status" db:"status"`
	Title        string            `json:"title" db:"title"`
	Description  string            `json:"description" db:"description"` // User's prompt/instructions
	Input        map[string]any    `json:"input,omitempty" db:"input"`   // Additional input parameters
	Output       map[string]any    `json:"output,omitempty" db:"output"` // Task results
	ErrorMessage *string           `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty" db:"completed_at"`
	Metadata     map[string]string `json:"metadata,omitempty" db:"metadata"`
	Tags         map[string]string `json:"tags,omitempty" db:"tags"`

	// Enhanced execution context (populated when requested)
	ExecutionContext TaskExecutionContext `json:"execution_context,omitempty" db:"-"`

	// Relationship data (populated when requested)
	Project  *Project  `json:"project,omitempty" db:"-"`
	Agent    *Agent    `json:"agent,omitempty" db:"-"`
	Codebase *Codebase `json:"codebase,omitempty" db:"-"`
}

// TaskExecutionContext holds essential context information about task execution
type TaskExecutionContext struct {
	// Execution environment details
	AgentVersion string `json:"agent_version,omitempty"`
	AIProvider   string `json:"ai_provider,omitempty"`
	ModelUsed    string `json:"model_used,omitempty"`

	// Basic performance metrics
	ExecutionTimeMs int64 `json:"execution_time_ms,omitempty"`
}

// CreateTaskRequest represents the request to create a new task
type CreateTaskRequest struct {
	ProjectID   string            `json:"project_id" validate:"required,project_id" example:"proj-12345-abcde"`
	AgentID     string            `json:"agent_id" validate:"required" example:"agent-12345"`
	CodebaseID  *string           `json:"codebase_id,omitempty" validate:"omitempty" example:"codebase-12345"`
	Type        TaskType          `json:"type" validate:"required,oneof=code_analysis refactoring code_review documentation custom" example:"refactoring"`
	Title       string            `json:"title" validate:"required,min=1,max=200" example:"Refactor authentication module"`
	Description string            `json:"description" validate:"required,min=1,max=2000" example:"Please refactor the user authentication module to use JWT tokens instead of sessions"`
	Input       map[string]any    `json:"input,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100"`
	Tags        map[string]string `json:"tags,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100"`
} //@name CreateTaskRequest

// CreateTaskResponse represents the response when creating a task
type CreateTaskResponse struct {
	TaskID    string     `json:"task_id" example:"task-12345-abcde"`
	Status    TaskStatus `json:"status" example:"pending"`
	CreatedAt time.Time  `json:"created_at" example:"2024-01-15T10:30:00Z"`
} //@name CreateTaskResponse

// GetTaskRequest represents the request to get a task by ID
type GetTaskRequest struct {
	TaskID string `uri:"id" validate:"required" example:"task-12345-abcde"`
} //@name GetTaskRequest

// GetTaskResponse represents the response when retrieving a task
type GetTaskResponse struct {
	Task
} //@name GetTaskResponse

// ListTasksRequest represents the request to list tasks for a project
type ListTasksRequest struct {
	ProjectID string      `uri:"project_id" validate:"required,project_id" example:"proj-12345-abcde"`
	Status    *TaskStatus `form:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed failed cancelled" example:"completed"`
	Type      *TaskType   `form:"type,omitempty" validate:"omitempty,oneof=code_analysis refactoring code_review documentation custom" example:"refactoring"`
	AgentID   *string     `form:"agent_id,omitempty" validate:"omitempty" example:"agent-12345"`
	Limit     *int        `form:"limit,omitempty" validate:"omitempty,min=1,max=100" example:"20"`
	Offset    *int        `form:"offset,omitempty" validate:"omitempty,min=0" example:"0"`
} //@name ListTasksRequest

// ListTasksResponse represents the response when listing tasks
type ListTasksResponse struct {
	Tasks      []Task `json:"tasks"`
	TotalCount int    `json:"total_count"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
} //@name ListTasksResponse

// UpdateTaskRequest represents the request to update a task
type UpdateTaskRequest struct {
	TaskID       string            `uri:"id" validate:"required" example:"task-12345-abcde"`
	Status       *TaskStatus       `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed failed cancelled"`
	Output       map[string]any    `json:"output,omitempty"`
	ErrorMessage *string           `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" validate:"omitempty,max=10,dive,keys,min=1,max=50,endkeys,min=1,max=100"`
} //@name UpdateTaskRequest

// UpdateTaskResponse represents the response when updating a task
type UpdateTaskResponse struct {
	Task
} //@name UpdateTaskResponse

// ExecuteTaskRequest represents the request to execute a task immediately
type ExecuteTaskRequest struct {
	ProjectID   string         `json:"project_id" validate:"required,project_id" example:"proj-12345-abcde"`
	AgentID     string         `json:"agent_id" validate:"required" example:"agent-12345"`
	CodebaseID  *string        `json:"codebase_id,omitempty" validate:"omitempty" example:"codebase-12345"`
	Type        TaskType       `json:"type" validate:"required,oneof=code_analysis refactoring code_review documentation custom" example:"refactoring"`
	Title       string         `json:"title" validate:"required,min=1,max=200" example:"Quick code analysis"`
	Description string         `json:"description" validate:"required,min=1,max=2000" example:"Analyze this function for potential improvements"`
	Input       map[string]any `json:"input,omitempty"`
	Async       bool           `json:"async" example:"false"` // If true, returns task ID; if false, waits for completion
} //@name ExecuteTaskRequest

// ExecuteTaskResponse represents the response when executing a task
type ExecuteTaskResponse struct {
	TaskID      string         `json:"task_id" example:"task-12345-abcde"`
	Status      TaskStatus     `json:"status" example:"completed"`
	Output      map[string]any `json:"output,omitempty"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-15T10:30:00Z"`
	CompletedAt *time.Time     `json:"completed_at,omitempty" example:"2024-01-15T10:35:00Z"`
} //@name ExecuteTaskResponse
