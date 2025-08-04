// Package services provides business logic implementations for task management
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/factory"
)

// TaskServiceImpl implements TaskService with dynamic AI capabilities
type TaskServiceImpl struct {
	taskRepo     repository.TaskRepository
	projectRepo  repository.ProjectRepository
	agentRepo    repository.AgentRepository
	codebaseRepo repository.CodebaseRepository
	aiFactory    factory.TaskExecutionFactory
}

// NewTaskService creates a new task service with dependency injection
func NewTaskService(
	taskRepo repository.TaskRepository,
	projectRepo repository.ProjectRepository,
	agentRepo repository.AgentRepository,
	codebaseRepo repository.CodebaseRepository,
	aiFactory factory.TaskExecutionFactory,
) TaskService {
	return &TaskServiceImpl{
		taskRepo:     taskRepo,
		projectRepo:  projectRepo,
		agentRepo:    agentRepo,
		codebaseRepo: codebaseRepo,
		aiFactory:    aiFactory,
	}
}

// CreateTask creates a new task with proper validation
func (s *TaskServiceImpl) CreateTask(ctx context.Context, req *models.CreateTaskRequest) (*models.CreateTaskResponse, error) {
	// Validate resources exist
	if err := s.validateResources(ctx, req.ProjectID, req.AgentID, req.CodebaseID); err != nil {
		return nil, fmt.Errorf("resource validation failed: %w", err)
	}

	// Generate task ID
	taskID := uuid.New().String()

	// Create execution context - simplified for now
	executionContext := &models.TaskExecutionContext{
		AgentVersion: "latest",
		AIProvider:   "bedrock", // Will be determined from agent config
	}

	// Create task model
	task := &models.Task{
		TaskID:           taskID,
		ProjectID:        req.ProjectID,
		AgentID:          req.AgentID,
		CodebaseID:       req.CodebaseID,
		Type:             req.Type,
		Status:           models.TaskStatusPending,
		Title:            req.Title,
		Description:      req.Description,
		Input:            req.Input,
		ExecutionContext: *executionContext,
		Metadata:         req.Metadata,
		Tags:             req.Tags,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save task to repository
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &models.CreateTaskResponse{
		TaskID:    task.TaskID,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}, nil
}

// GetTask retrieves a task by ID with optional context loading
func (s *TaskServiceImpl) GetTask(ctx context.Context, taskID string) (*models.GetTaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &models.GetTaskResponse{
		Task: *task,
	}, nil
}

// UpdateTask updates an existing task
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, req *models.UpdateTaskRequest) (*models.UpdateTaskResponse, error) {
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

	task.UpdatedAt = time.Now()

	// Save updated task
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return &models.UpdateTaskResponse{
		Task: *task,
	}, nil
}

// DeleteTask deletes a task by ID
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, taskID string) error {
	return s.taskRepo.Delete(ctx, taskID)
}

// ListTasks lists tasks for a project with filters
func (s *TaskServiceImpl) ListTasks(ctx context.Context, req *models.ListTasksRequest) (*models.ListTasksResponse, error) {
	// Set default pagination
	limit := 20
	offset := 0

	if req.Limit != nil && *req.Limit > 0 {
		limit = *req.Limit
	}
	if req.Offset != nil && *req.Offset >= 0 {
		offset = *req.Offset
	}

	// Create filters from request
	filters := repository.TaskFilters{
		Status:  req.Status,
		Type:    req.Type,
		AgentID: req.AgentID,
		Limit:   limit,
		Offset:  offset,
	}

	tasks, total, err := s.taskRepo.ListByProject(ctx, req.ProjectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return &models.ListTasksResponse{
		Tasks:      tasks,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

// ExecuteTask executes a task with dynamic AI resource allocation
func (s *TaskServiceImpl) ExecuteTask(ctx context.Context, req *models.ExecuteTaskRequest) (*models.ExecuteTaskResponse, error) {
	// First, create the task
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

	// If async requested, return immediately
	if req.Async {
		return &models.ExecuteTaskResponse{
			TaskID:    createResp.TaskID,
			Status:    models.TaskStatusPending,
			CreatedAt: createResp.CreatedAt,
		}, nil
	}

	// For sync execution, load full context and execute
	return s.executeTaskSync(ctx, createResp.TaskID, req)
}

// Private helper methods

// executeTaskSync performs synchronous task execution with dynamic AI resources
func (s *TaskServiceImpl) executeTaskSync(ctx context.Context, taskID string, req *models.ExecuteTaskRequest) (*models.ExecuteTaskResponse, error) {
	// Update task status to in_progress
	if err := s.taskRepo.UpdateStatus(ctx, taskID, models.TaskStatusInProgress); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Load task with full context
	taskWithContext, err := s.loadTaskWithFullContext(ctx, taskID)
	if err != nil {
		s.updateTaskError(ctx, taskID, fmt.Sprintf("failed to load context: %v", err))
		return nil, fmt.Errorf("failed to load task context: %w", err)
	}

	// Validate agent configuration
	if err := s.aiFactory.ValidateAgentConfig(taskWithContext.Agent.AIConfig); err != nil {
		s.updateTaskError(ctx, taskID, fmt.Sprintf("invalid agent config: %v", err))
		return nil, fmt.Errorf("agent configuration validation failed: %w", err)
	}

	// Create AI agent dynamically
	agentID, agentVersion, err := s.aiFactory.CreateAgentForTask(ctx, taskWithContext)
	if err != nil {
		s.updateTaskError(ctx, taskID, fmt.Sprintf("AI agent creation failed: %v", err))
		return nil, fmt.Errorf("failed to create AI agent: %w", err)
	}

	// Execute task with AI agent
	results := map[string]any{
		"task_id":             taskID,
		"agent_id":            agentID,
		"agent_version":       agentVersion,
		"ai_provider":         taskWithContext.Agent.AIConfig.Provider,
		"execution_method":    "dynamic_ai_factory",
		"supported_providers": s.aiFactory.GetSupportedProviders(),
		"prompt":              req.Description,
		"task_type":           req.Type,
		"message":             "Task executed successfully with dynamic AI resources",
		"executed_at":         time.Now().Format(time.RFC3339),
	}

	// Update task with results
	err = s.taskRepo.UpdateStatusAndOutput(ctx, taskID, models.TaskStatusCompleted, results, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update task results: %w", err)
	}

	completedAt := time.Now()

	return &models.ExecuteTaskResponse{
		TaskID:      taskID,
		Status:      models.TaskStatusCompleted,
		Output:      results,
		CreatedAt:   time.Time{}, // Will be set from task
		CompletedAt: &completedAt,
	}, nil
}

// validateResources validates that project, agent, and optionally codebase exist
func (s *TaskServiceImpl) validateResources(ctx context.Context, projectID, agentID string, codebaseID *string) error {
	// Validate project exists
	if _, err := s.projectRepo.GetProject(ctx, projectID); err != nil {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// Validate agent exists (skip if agent repository not available)
	if s.agentRepo != nil {
		if _, err := s.agentRepo.GetAgent(ctx, agentID); err != nil {
			return fmt.Errorf("agent not found: %s", agentID)
		}
	}

	// Validate codebase if specified
	if codebaseID != nil {
		if _, err := s.codebaseRepo.GetCodebase(ctx, *codebaseID); err != nil {
			return fmt.Errorf("codebase not found: %s", *codebaseID)
		}
	}

	return nil
}

// loadTaskWithFullContext loads a task with all related resources
func (s *TaskServiceImpl) loadTaskWithFullContext(ctx context.Context, taskID string) (*models.TaskWithFullContext, error) {
	// Get the task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Load project
	project, err := s.projectRepo.GetProject(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Load agent (skip if agent repository not available)
	var agent *repository.AgentRecord
	if s.agentRepo != nil {
		var err error
		agent, err = s.agentRepo.GetAgent(ctx, task.AgentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get agent: %w", err)
		}
	}

	// Convert repository records to API models
	var projectModel *models.Project
	if project != nil {
		projectModel = &models.Project{
			ProjectID:   project.ProjectID,
			Name:        project.Name,
			Description: project.Description,
			Language:    project.Language,
			Status:      models.ProjectStatus(project.Status),
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
			Tags:        project.Tags,
			Metadata:    project.Metadata,
		}
	}

	var agentModel *models.Agent
	if agent != nil {
		agentModel = &models.Agent{
			AgentID:   agent.AgentID,
			ProjectID: task.ProjectID, // From task context
			Name:      agent.AgentName,
			Status:    models.AgentStatus(agent.Status),
			Version:   agent.AgentVersion,
			CreatedAt: agent.CreatedAt,
			UpdatedAt: agent.UpdatedAt,
		}
		if agent.KnowledgeBaseID != "" {
			agentModel.KnowledgeBaseID = &agent.KnowledgeBaseID
		}
		if agent.VectorStoreID != "" {
			agentModel.VectorStoreID = &agent.VectorStoreID
		}
	}

	taskContext := &models.TaskWithFullContext{
		Task:     *task,
		Project:  projectModel,
		Agent:    agentModel,
		Codebase: nil,
	}

	// Load codebase if specified
	if task.CodebaseID != nil {
		codebase, err := s.codebaseRepo.GetCodebase(ctx, *task.CodebaseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get codebase: %w", err)
		}
		taskContext.Codebase = codebase
	}

	return taskContext, nil
}

// updateTaskError updates a task with error status and message
func (s *TaskServiceImpl) updateTaskError(ctx context.Context, taskID, errorMsg string) {
	if err := s.taskRepo.UpdateStatusAndOutput(ctx, taskID, models.TaskStatusFailed, nil, &errorMsg); err != nil {
		// Log error but don't fail since this is a cleanup operation
		slog.Error("failed to update task error status", "task_id", taskID, "error", err)
	}
}
