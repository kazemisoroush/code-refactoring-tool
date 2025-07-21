// Package models provides data structures for API requests and responses
package models

import (
	"time"
)

// CreateAgentRequest represents the request to create a new agent
type CreateAgentRequest struct {
	// Repository URL to analyze
	RepositoryURL string `json:"repository_url" binding:"required" example:"https://github.com/user/repo"`
	// Optional branch name, defaults to main
	Branch string `json:"branch,omitempty" example:"main"`
	// Optional custom agent name
	AgentName string `json:"agent_name,omitempty" example:"my-code-analyzer"`
} //@name CreateAgentRequest

// CreateAgentResponse represents the response when creating an agent
type CreateAgentResponse struct {
	// Unique identifier for the agent
	AgentID string `json:"agent_id" example:"agent-12345"`
	// Agent version
	AgentVersion string `json:"agent_version" example:"v1.0.0"`
	// Knowledge base ID associated with the agent
	KnowledgeBaseID string `json:"knowledge_base_id" example:"kb-67890"`
	// Vector store ID for the knowledge base
	VectorStoreID string `json:"vector_store_id" example:"vs-abcde"`
	// Agent creation status
	Status string `json:"status" example:"created"`
	// Timestamp when the agent was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
} //@name CreateAgentResponse

// AgentStatus represents the status of an agent
type AgentStatus string

// Agent status constants
const (
	// AgentStatusCreating indicates the agent is being created
	AgentStatusCreating AgentStatus = "creating"
	// AgentStatusReady indicates the agent is ready for use
	AgentStatusReady AgentStatus = "ready"
	// AgentStatusError indicates the agent encountered an error
	AgentStatusError AgentStatus = "error"
	// AgentStatusDeleted indicates the agent has been deleted
	AgentStatusDeleted AgentStatus = "deleted"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Error code
	Code int `json:"code" example:"400"`
	// Error message
	Message string `json:"message" example:"Invalid request parameters"`
	// Optional error details
	Details string `json:"details,omitempty" example:"repository_url is required"`
} //@name ErrorResponse
