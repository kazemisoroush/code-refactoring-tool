// Package models provides data structures for API requests and responses
package models

import (
	"time"
)

// CreateAgentRequest represents the request to create a new agent
type CreateAgentRequest struct {
	// Repository URL to analyze
	RepositoryURL string `json:"repository_url" binding:"required" validate:"required,url" example:"https://github.com/user/repo"`
	// Optional branch name, defaults to main
	Branch string `json:"branch,omitempty" validate:"omitempty,min=1" example:"main"`
	// Optional custom agent name
	AgentName string `json:"agent_name,omitempty" validate:"omitempty,min=1" example:"my-code-analyzer"`
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

// GetAgentRequest represents the request to get an agent by ID
type GetAgentRequest struct {
	// Agent ID (UUID)
	AgentID string `uri:"id" validate:"required" example:"agent-12345"`
} //@name GetAgentRequest

// GetAgentResponse represents the response when retrieving an agent
type GetAgentResponse struct {
	// Unique identifier for the agent
	AgentID string `json:"agent_id" example:"agent-12345"`
	// Agent version
	AgentVersion string `json:"agent_version" example:"v1.0.0"`
	// Knowledge base ID associated with the agent
	KnowledgeBaseID string `json:"knowledge_base_id" example:"kb-67890"`
	// Vector store ID for the knowledge base
	VectorStoreID string `json:"vector_store_id" example:"vs-abcde"`
	// Repository URL
	RepositoryURL string `json:"repository_url" example:"https://github.com/user/repo"`
	// Branch name
	Branch string `json:"branch" example:"main"`
	// Agent name
	AgentName string `json:"agent_name" example:"my-code-analyzer"`
	// Agent status
	Status string `json:"status" example:"ready"`
	// Timestamp when the agent was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when the agent was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
} //@name GetAgentResponse

// DeleteAgentRequest represents the request to delete an agent by ID
type DeleteAgentRequest struct {
	// Agent ID
	AgentID string `uri:"id" validate:"required" example:"agent-12345"`
} //@name DeleteAgentRequest

// DeleteAgentResponse represents the response when deleting an agent
type DeleteAgentResponse struct {
	// Agent ID that was deleted
	AgentID string `json:"agent_id" example:"agent-12345"`
	// Success indicator
	Success bool `json:"success" example:"true"`
} //@name DeleteAgentResponse

// ListAgentsRequest represents the request to list agents
type ListAgentsRequest struct {
	// Token for pagination
	NextToken *string `form:"next_token" validate:"omitempty" example:"eyJ0aW1lc3RhbXAiOiIyMDI0LTAxLTE1VDEwOjMwOjAwWiJ9"`
	// Maximum number of results to return
	MaxResults *int `form:"max_results" validate:"omitempty,min=1,max=100" example:"20"`
} //@name ListAgentsRequest

// AgentSummary represents a summary of an agent for listing
type AgentSummary struct {
	// Unique identifier for the agent
	AgentID string `json:"agent_id" example:"agent-12345"`
	// Agent name
	AgentName string `json:"agent_name" example:"my-code-analyzer"`
	// Repository URL
	RepositoryURL string `json:"repository_url" example:"https://github.com/user/repo"`
	// Agent status
	Status string `json:"status" example:"ready"`
	// Timestamp when the agent was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
} //@name AgentSummary

// ListAgentsResponse represents the response when listing agents
type ListAgentsResponse struct {
	// List of agent summaries
	Agents []AgentSummary `json:"agents"`
	// Token for next page
	NextToken string `json:"next_token,omitempty" example:"eyJ0aW1lc3RhbXAiOiIyMDI0LTAxLTE1VDEwOjMwOjAwWiJ9"`
} //@name ListAgentsResponse

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
