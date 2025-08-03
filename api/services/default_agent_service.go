// Package services provides concrete implementations for business logic
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	pkgRepo "github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

// DefaultAgentService is the default implementation of AgentService
type DefaultAgentService struct {
	gitConfig       config.GitConfig
	ragBuilder      builder.RAGBuilder
	agentBuilder    builder.AgentBuilder
	gitRepository   pkgRepo.Codebase
	agentRepository repository.AgentRepository
}

// NewDefaultAgentService creates a new instance of DefaultAgentService
func NewDefaultAgentService(
	gitConfig config.GitConfig,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
	gitRepo pkgRepo.Codebase,
	agentRepo repository.AgentRepository,
) AgentService {
	return &DefaultAgentService{
		gitConfig:       gitConfig,
		ragBuilder:      ragBuilder,
		agentBuilder:    agentBuilder,
		gitRepository:   gitRepo,
		agentRepository: agentRepo,
	}
}

// CreateAgent creates a new agent with the given parameters
func (s *DefaultAgentService) CreateAgent(ctx context.Context, request models.CreateAgentRequest) (*models.CreateAgentResponse, error) {
	slog.Info("Creating agent", "repository_url", request.RepositoryURL, "branch", request.Branch)

	// Create a new repository instance for this request
	repoConfig := s.gitConfig
	repoConfig.CodebaseURL = request.RepositoryURL
	// Note: Branch handling would need to be implemented in repository layer if needed

	repo := pkgRepo.NewGitHubCodebase(repoConfig)

	// Create and run setup workflow
	setupWorkflow, err := workflow.NewSetupWorkflow(repo, s.ragBuilder, s.agentBuilder)
	if err != nil {
		slog.Error("Failed to create setup workflow", "error", err)
		return nil, fmt.Errorf("failed to create setup workflow: %w", err)
	}

	err = setupWorkflow.Run(ctx)
	if err != nil {
		slog.Error("Failed to run setup workflow", "error", err)
		return nil, fmt.Errorf("failed to setup agent: %w", err)
	}

	// Extract resource IDs from the workflow
	// Cast to concrete type to access GetResourceIDs method
	concreteWorkflow := setupWorkflow.(*workflow.SetupWorkflow)
	vectorStoreID, ragID, agentID, agentVersion := concreteWorkflow.GetResourceIDs()

	// Create agent record for storage
	agentRecord := repository.NewAgentRecord(request, agentID, agentVersion, ragID, vectorStoreID)

	// Store the agent record in DynamoDB
	err = s.agentRepository.CreateAgent(ctx, agentRecord)
	if err != nil {
		slog.Error("Failed to store agent record", "error", err)
		// Note: In production, we might want to trigger cleanup of AWS resources here
		return nil, fmt.Errorf("failed to store agent record: %w", err)
	}

	response := agentRecord.ToResponse()

	slog.Info("Agent created successfully",
		"agent_id", agentID,
		"agent_version", agentVersion,
		"knowledge_base_id", ragID,
		"vector_store_id", vectorStoreID)

	return response, nil
}

// GetAgent retrieves an agent by ID
func (s *DefaultAgentService) GetAgent(ctx context.Context, agentID string) (*models.GetAgentResponse, error) {
	agentRecord, err := s.agentRepository.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	response := &models.GetAgentResponse{
		AgentID:         agentRecord.AgentID,
		AgentVersion:    agentRecord.AgentVersion,
		KnowledgeBaseID: agentRecord.KnowledgeBaseID,
		VectorStoreID:   agentRecord.VectorStoreID,
		RepositoryURL:   agentRecord.RepositoryURL,
		Branch:          agentRecord.Branch,
		AgentName:       agentRecord.AgentName,
		Status:          agentRecord.Status,
		CreatedAt:       agentRecord.CreatedAt,
		UpdatedAt:       agentRecord.UpdatedAt,
	}

	return response, nil
}

// DeleteAgent deletes an agent by ID
func (s *DefaultAgentService) DeleteAgent(ctx context.Context, agentID string) (*models.DeleteAgentResponse, error) {
	// First get the agent to retrieve resource IDs for cleanup
	agentRecord, err := s.agentRepository.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent for deletion: %w", err)
	}

	// Update status to deleting
	err = s.agentRepository.UpdateAgentStatus(ctx, agentID, models.AgentStatusDeleted)
	if err != nil {
		slog.Error("Failed to update agent status to deleting", "error", err)
		// Continue with deletion anyway
	}

	// Create teardown workflow to clean up AWS resources
	teardownWorkflow, err := workflow.NewTeardownWorkflowWithResources(
		s.gitRepository, // Use a dummy repository for teardown
		s.ragBuilder,
		s.agentBuilder,
		agentRecord.VectorStoreID,
		agentRecord.KnowledgeBaseID,
		agentRecord.AgentID,
		agentRecord.AgentVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create teardown workflow: %w", err)
	}

	// Run teardown workflow
	err = teardownWorkflow.Run(ctx)
	if err != nil {
		slog.Error("Failed to run teardown workflow", "error", err)
		// Don't return error here - we still want to delete from DB
	}

	// Remove from database
	err = s.agentRepository.DeleteAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete agent from database: %w", err)
	}

	response := &models.DeleteAgentResponse{
		AgentID: agentID,
		Success: true,
	}

	slog.Info("Agent deleted successfully", "agent_id", agentID)
	return response, nil
}

// ListAgents lists all agents with optional pagination
func (s *DefaultAgentService) ListAgents(ctx context.Context, request models.ListAgentsRequest) (*models.ListAgentsResponse, error) {
	agentRecords, err := s.agentRepository.ListAgents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	// Convert to summary format
	summaries := make([]models.AgentSummary, len(agentRecords))
	for i, record := range agentRecords {
		summaries[i] = models.AgentSummary{
			AgentID:       record.AgentID,
			AgentName:     record.AgentName,
			RepositoryURL: record.RepositoryURL,
			Status:        record.Status,
			CreatedAt:     record.CreatedAt,
		}
	}

	// For now, we'll implement basic pagination logic (can be enhanced later)
	maxResults := 50 // default
	if request.MaxResults != nil {
		maxResults = *request.MaxResults
	}

	// Simple pagination: return all for now, but structure for future enhancement
	response := &models.ListAgentsResponse{
		Agents: summaries,
	}

	// If we have more results than max, we'd set NextToken here
	if len(summaries) > maxResults {
		response.Agents = summaries[:maxResults]
		response.NextToken = "next_page_token" // Placeholder
	}

	return response, nil
}
