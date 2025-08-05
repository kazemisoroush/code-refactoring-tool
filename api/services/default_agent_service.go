// Package services provides concrete implementations for business logic
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/factory"
)

// DefaultAgentService is the default implementation of AgentService
type DefaultAgentService struct {
	agentRepository       repository.AgentRepository
	infrastructureFactory factory.AIInfrastructureFactory
}

// NewDefaultAgentService creates a new instance of DefaultAgentService
func NewDefaultAgentService(
	agentRepo repository.AgentRepository,
	infraFactory factory.AIInfrastructureFactory,
) AgentService {
	return &DefaultAgentService{
		agentRepository:       agentRepo,
		infrastructureFactory: infraFactory,
	}
}

// CreateAgent creates a new agent with the given parameters
func (s *DefaultAgentService) CreateAgent(ctx context.Context, request models.CreateAgentRequest) (*models.CreateAgentResponse, error) {
	slog.Info("Creating agent", "repository_url", request.RepositoryURL, "branch", request.Branch)

	// For now, we'll use a default AI configuration since the request doesn't include one
	// TODO: Update CreateAgentRequest to include AI configuration in a future iteration
	defaultAIConfig := &models.AgentAIConfig{
		Provider: models.AIProviderLocal, // Default to local for simplicity
		Local: &models.LocalAgentConfig{
			OllamaURL: "http://localhost:11434", // Default Ollama URL
			Model:     "llama3.1:latest",        // Default model
		},
	}

	// Validate the AI configuration
	if err := s.infrastructureFactory.ValidateAgentConfig(defaultAIConfig); err != nil {
		return nil, fmt.Errorf("invalid AI configuration: %w", err)
	}

	// Create AI infrastructure
	infraResult, err := s.infrastructureFactory.CreateAgentInfrastructure(ctx, defaultAIConfig, request.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI infrastructure: %w", err)
	}

	// Create agent record to store in database
	agentRecord := &repository.AgentRecord{
		AgentID:         infraResult.AgentID,
		AgentVersion:    infraResult.AgentVersion,
		KnowledgeBaseID: infraResult.KnowledgeBaseID,
		VectorStoreID:   infraResult.VectorStoreID,
		RepositoryURL:   request.RepositoryURL,
		Branch:          request.Branch,
		AgentName:       request.AgentName,
		Status:          string(infraResult.Status),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Set default agent name if not provided
	if agentRecord.AgentName == "" {
		agentRecord.AgentName = fmt.Sprintf("agent-%s", infraResult.AgentID[:8])
	}

	// Set default branch if not provided
	if agentRecord.Branch == "" {
		agentRecord.Branch = "main"
	}

	// Save agent to repository
	err = s.agentRepository.CreateAgent(ctx, agentRecord)
	if err != nil {
		// If saving fails, try to clean up the infrastructure
		cleanupErr := s.infrastructureFactory.DestroyAgentInfrastructure(ctx, infraResult.AgentID)
		if cleanupErr != nil {
			slog.Error("Failed to cleanup infrastructure after database save failure",
				"agent_id", infraResult.AgentID, "cleanup_error", cleanupErr)
		}
		return nil, fmt.Errorf("failed to save agent to database: %w", err)
	}

	// Update status to ready after successful creation
	err = s.agentRepository.UpdateAgentStatus(ctx, infraResult.AgentID, models.AgentStatusReady)
	if err != nil {
		slog.Error("Failed to update agent status to ready", "agent_id", infraResult.AgentID, "error", err)
		// Don't fail the creation for this, just log it
	}

	response := &models.CreateAgentResponse{
		AgentID:         infraResult.AgentID,
		AgentVersion:    infraResult.AgentVersion,
		KnowledgeBaseID: infraResult.KnowledgeBaseID,
		VectorStoreID:   infraResult.VectorStoreID,
		Status:          string(models.AgentStatusReady),
		CreatedAt:       agentRecord.CreatedAt,
	}

	slog.Info("Agent created successfully", "agent_id", infraResult.AgentID)
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

// UpdateAgent updates an existing agent
func (s *DefaultAgentService) UpdateAgent(ctx context.Context, request models.UpdateAgentRequest) (*models.UpdateAgentResponse, error) {
	slog.Info("Updating agent", "agent_id", request.AgentID)

	// Get existing agent to ensure it exists
	existingAgent, err := s.agentRepository.GetAgent(ctx, request.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing agent: %w", err)
	}

	// Create update record with current values as defaults
	updateRecord := &repository.AgentRecord{
		AgentID:       existingAgent.AgentID,
		AgentName:     existingAgent.AgentName,
		RepositoryURL: existingAgent.RepositoryURL,
		Branch:        existingAgent.Branch,
		Status:        existingAgent.Status,
	}

	// Apply updates if provided
	if request.AgentName != nil {
		updateRecord.AgentName = *request.AgentName
	}
	if request.RepositoryURL != nil {
		updateRecord.RepositoryURL = *request.RepositoryURL
	}
	if request.Branch != nil {
		updateRecord.Branch = *request.Branch
	}

	// Update the agent in the repository
	err = s.agentRepository.UpdateAgent(ctx, updateRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	// Get the updated agent to return current state
	updatedAgent, err := s.agentRepository.GetAgent(ctx, request.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated agent: %w", err)
	}

	response := &models.UpdateAgentResponse{
		AgentID:         updatedAgent.AgentID,
		AgentVersion:    updatedAgent.AgentVersion,
		KnowledgeBaseID: updatedAgent.KnowledgeBaseID,
		VectorStoreID:   updatedAgent.VectorStoreID,
		RepositoryURL:   updatedAgent.RepositoryURL,
		Branch:          updatedAgent.Branch,
		AgentName:       updatedAgent.AgentName,
		Status:          updatedAgent.Status,
		CreatedAt:       updatedAgent.CreatedAt,
		UpdatedAt:       updatedAgent.UpdatedAt,
	}

	return response, nil
}

// DeleteAgent deletes an agent by ID
func (s *DefaultAgentService) DeleteAgent(ctx context.Context, agentID string) (*models.DeleteAgentResponse, error) {
	// Verify agent exists before attempting deletion
	_, err := s.agentRepository.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent for deletion: %w", err)
	}

	// Update status to deleting
	err = s.agentRepository.UpdateAgentStatus(ctx, agentID, models.AgentStatusDeleted)
	if err != nil {
		slog.Error("Failed to update agent status to deleting", "error", err)
		// Continue with deletion anyway
	}

	// Use the infrastructure factory to clean up AI resources
	err = s.infrastructureFactory.DestroyAgentInfrastructure(ctx, agentID)
	if err != nil {
		slog.Error("Failed to destroy AI infrastructure", "agent_id", agentID, "error", err)
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
