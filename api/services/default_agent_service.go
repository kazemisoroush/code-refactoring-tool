// Package services provides concrete implementations for business logic
package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
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

	// Use provided AI provider or default to local
	aiProvider := request.AIProvider
	if aiProvider == "" {
		aiProvider = models.AIProviderLocal
	}

	if err := s.infrastructureFactory.ValidateAgentConfig(aiProvider); err != nil {
		return nil, fmt.Errorf("invalid AI provider: %w", err)
	}

	// Create AI infrastructure
	infraResult, err := s.infrastructureFactory.CreateAgentInfrastructure(ctx, aiProvider, request.RepositoryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI infrastructure: %w", err)
	}

	// Create agent record to store in database
	agentRecord := repository.NewAgentRecord(request, infraResult.AgentID, infraResult.AgentVersion, infraResult.KnowledgeBaseID, infraResult.VectorStoreID)
	agentRecord.Status = string(infraResult.Status)
	agentRecord.CreatedAt = time.Now()
	agentRecord.UpdatedAt = time.Now()

	// Set default agent name if not provided
	if agentRecord.AgentName == "" {
		agentRecord.AgentName = fmt.Sprintf("agent-%s", infraResult.AgentID[:8])
	}

	// Set default branch if not provided
	if agentRecord.Branch == "" {
		agentRecord.Branch = config.DefaultGitBranch
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

	// Check if infrastructure-impacting changes are being made
	infrastructureChangeRequired := false
	newRepositoryURL := existingAgent.RepositoryURL

	if request.RepositoryURL != nil && *request.RepositoryURL != existingAgent.RepositoryURL {
		infrastructureChangeRequired = true
		newRepositoryURL = *request.RepositoryURL
	}
	if request.Branch != nil && *request.Branch != existingAgent.Branch {
		infrastructureChangeRequired = true
	}
	if request.AIProvider != nil {
		infrastructureChangeRequired = true
	}

	// If infrastructure changes are required, update the AI infrastructure
	var infrastructureResult *factory.AIInfrastructureResult
	if infrastructureChangeRequired {
		slog.Info("Infrastructure changes detected, updating AI infrastructure", "agent_id", request.AgentID)

		// Use provided AI provider or keep existing agent's provider
		var aiProvider models.AIProvider
		if request.AIProvider != nil {
			aiProvider = *request.AIProvider
		} else {
			aiProvider = models.AIProvider(existingAgent.AIProvider)
		}

		infrastructureResult, err = s.infrastructureFactory.UpdateAgentInfrastructure(ctx, existingAgent.KnowledgeBaseID, aiProvider, newRepositoryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to update AI infrastructure: %w", err)
		}
	}

	// Create update record with current values as defaults
	updateRecord := &repository.AgentRecord{
		AgentID:       existingAgent.AgentID,
		AgentName:     existingAgent.AgentName,
		RepositoryURL: existingAgent.RepositoryURL,
		Branch:        existingAgent.Branch,
		Status:        existingAgent.Status,
		AIProvider:    existingAgent.AIProvider,
		AIConfigJSON:  existingAgent.AIConfigJSON,
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
	if request.AIProvider != nil {
		updateRecord.AIProvider = string(*request.AIProvider)
	}

	// If infrastructure was recreated, update the infrastructure IDs
	if infrastructureResult != nil {
		updateRecord.KnowledgeBaseID = infrastructureResult.KnowledgeBaseID
		updateRecord.VectorStoreID = infrastructureResult.VectorStoreID
		updateRecord.AgentVersion = infrastructureResult.AgentVersion
		updateRecord.Status = string(infrastructureResult.Status)
	}

	// Update the agent in the repository
	err = s.agentRepository.UpdateAgent(ctx, updateRecord)
	if err != nil {
		// If database update fails but we updated infrastructure, we should try to revert
		if infrastructureResult != nil {
			slog.Warn("Database update failed after updating infrastructure, attempting to revert",
				"agent_id", request.AgentID, "new_knowledge_base_id", infrastructureResult.KnowledgeBaseID)
			// Note: Reverting infrastructure updates is complex and may not always be possible
			// In a production system, you might want to implement a compensation transaction
		}
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
