// Package repository provides data access layer for the API
package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// AgentRecord represents the agent data stored in the database
type AgentRecord struct {
	AgentID         string    `json:"agent_id" db:"agent_id"`
	AgentVersion    string    `json:"agent_version" db:"agent_version"`
	KnowledgeBaseID string    `json:"knowledge_base_id" db:"knowledge_base_id"`
	VectorStoreID   string    `json:"vector_store_id" db:"vector_store_id"`
	RepositoryURL   string    `json:"repository_url" db:"repository_url"`
	Branch          string    `json:"branch,omitempty" db:"branch"`
	AgentName       string    `json:"agent_name,omitempty" db:"agent_name"`
	Status          string    `json:"status" db:"status"`
	AIProvider      string    `json:"ai_provider,omitempty" db:"ai_provider"`
	AIConfigJSON    string    `json:"ai_config_json,omitempty" db:"ai_config_json"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// ToResponse converts AgentRecord to CreateAgentResponse
func (r *AgentRecord) ToResponse() *models.CreateAgentResponse {
	return &models.CreateAgentResponse{
		AgentID:         r.AgentID,
		AgentVersion:    r.AgentVersion,
		KnowledgeBaseID: r.KnowledgeBaseID,
		VectorStoreID:   r.VectorStoreID,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
	}
}

// NewAgentRecord creates an AgentRecord from CreateAgentRequest
func NewAgentRecord(request models.CreateAgentRequest, agentID, agentVersion, kbID, vectorStoreID string) *AgentRecord {
	now := time.Now().UTC()

	record := &AgentRecord{
		AgentID:         agentID,
		AgentVersion:    agentVersion,
		KnowledgeBaseID: kbID,
		VectorStoreID:   vectorStoreID,
		RepositoryURL:   request.RepositoryURL,
		Branch:          request.Branch,
		AgentName:       request.AgentName,
		Status:          string(models.AgentStatusReady),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Set AI config if provided
	if request.AIConfig != nil {
		record.AIProvider = string(request.AIConfig.Provider)

		// Serialize AI config to JSON for storage
		if configJSON, err := json.Marshal(request.AIConfig); err == nil {
			record.AIConfigJSON = string(configJSON)
		}
	}

	return record
}

// GetAIConfig deserializes the stored AI configuration
func (r *AgentRecord) GetAIConfig() (*models.AgentAIConfig, error) {
	if r.AIConfigJSON == "" {
		return nil, nil
	}

	var config models.AgentAIConfig
	if err := json.Unmarshal([]byte(r.AIConfigJSON), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// AgentRepository defines the interface for agent data operations
//
//go:generate mockgen -destination=./mocks/mock_agent_repository.go -mock_names=AgentRepository=MockAgentRepository -package=mocks . AgentRepository
type AgentRepository interface {
	// CreateAgent stores a new agent record
	CreateAgent(ctx context.Context, agent *AgentRecord) error

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, agentID string) (*AgentRecord, error)

	// UpdateAgent updates an existing agent record
	UpdateAgent(ctx context.Context, agent *AgentRecord) error

	// DeleteAgent removes an agent record
	DeleteAgent(ctx context.Context, agentID string) error

	// ListAgents retrieves all agent records
	ListAgents(ctx context.Context) ([]*AgentRecord, error)

	// UpdateAgentStatus updates only the status field
	UpdateAgentStatus(ctx context.Context, agentID string, status models.AgentStatus) error
}
