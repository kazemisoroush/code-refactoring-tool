// Package services provides business logic for the API layer
package services

import (
	"context"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// AgentService defines the interface for agent-related operations
//
//go:generate mockgen -destination=./mocks/mock_agent_service.go -mock_names=AgentService=MockAgentService -package=mocks . AgentService
type AgentService interface {
	// CreateAgent creates a new agent with the given parameters
	CreateAgent(ctx context.Context, request models.CreateAgentRequest) (*models.CreateAgentResponse, error)

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, agentID string) (*models.CreateAgentResponse, error)

	// DeleteAgent deletes an agent by ID
	DeleteAgent(ctx context.Context, agentID string) error

	// ListAgents lists all agents
	ListAgents(ctx context.Context) ([]*models.CreateAgentResponse, error)
}
