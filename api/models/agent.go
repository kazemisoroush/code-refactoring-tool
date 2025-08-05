// Package models provides data structures for API requests and responses
package models

import (
	"time"
)

// Agent represents a complete agent with its AI configuration
type Agent struct {
	AgentID     string      `json:"agent_id" db:"agent_id"`
	ProjectID   string      `json:"project_id" db:"project_id"`
	Name        string      `json:"name" db:"name"`
	Description *string     `json:"description,omitempty" db:"description"`
	Status      AgentStatus `json:"status" db:"status"`

	// AI Configuration - this is per-agent, not global
	AIConfig AgentAIConfig `json:"ai_config" db:"ai_config"`

	// Agent capabilities and metadata
	Capabilities []string `json:"capabilities" db:"capabilities"` // e.g., ["code_analysis", "refactoring", "review"]
	Version      string   `json:"version" db:"version"`

	// Knowledge base and vector store IDs (provider-specific)
	KnowledgeBaseID *string `json:"knowledge_base_id,omitempty" db:"knowledge_base_id"`
	VectorStoreID   *string `json:"vector_store_id,omitempty" db:"vector_store_id"`

	// Timestamps and metadata
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at" db:"updated_at"`
	LastUsedAt *time.Time        `json:"last_used_at,omitempty" db:"last_used_at"`
	Metadata   map[string]string `json:"metadata,omitempty" db:"metadata"`
	Tags       map[string]string `json:"tags,omitempty" db:"tags"`
}

// AgentAIConfig represents AI configuration specific to an agent
type AgentAIConfig struct {
	Provider AIProvider `json:"provider" db:"provider"` // bedrock, local, openai

	// Provider-specific configurations
	Local   *LocalAgentConfig   `json:"local,omitempty" db:"local"`
	Bedrock *BedrockAgentConfig `json:"bedrock,omitempty" db:"bedrock"`
	OpenAI  *OpenAIAgentConfig  `json:"openai,omitempty" db:"openai"`
}

// LocalAgentConfig represents local AI configuration for an agent
type LocalAgentConfig struct {
	OllamaURL      string  `json:"ollama_url" db:"ollama_url"`
	Model          string  `json:"model" db:"model"`
	ChromaURL      string  `json:"chroma_url" db:"chroma_url"`
	EmbeddingModel string  `json:"embedding_model" db:"embedding_model"`
	Temperature    float64 `json:"temperature,omitempty" db:"temperature"`
	MaxTokens      int     `json:"max_tokens,omitempty" db:"max_tokens"`
}

// BedrockAgentConfig represents AWS Bedrock configuration for an agent
type BedrockAgentConfig struct {
	Region                      string  `json:"region" db:"region"`
	FoundationModel             string  `json:"foundation_model" db:"foundation_model"`
	EmbeddingModel              string  `json:"embedding_model" db:"embedding_model"`
	KnowledgeBaseServiceRoleARN string  `json:"knowledge_base_service_role_arn" db:"knowledge_base_service_role_arn"`
	AgentServiceRoleARN         string  `json:"agent_service_role_arn" db:"agent_service_role_arn"`
	S3BucketName                string  `json:"s3_bucket_name" db:"s3_bucket_name"`
	Temperature                 float64 `json:"temperature,omitempty" db:"temperature"`
	MaxTokens                   int     `json:"max_tokens,omitempty" db:"max_tokens"`
}

// OpenAIAgentConfig represents OpenAI configuration for an agent (future extension)
type OpenAIAgentConfig struct {
	APIKey      string  `json:"api_key" db:"api_key"`
	Model       string  `json:"model" db:"model"`
	Temperature float64 `json:"temperature,omitempty" db:"temperature"`
	MaxTokens   int     `json:"max_tokens,omitempty" db:"max_tokens"`
}

// CreateAgentRequest represents the request to create a new agent
type CreateAgentRequest struct {
	// Repository URL to analyze
	RepositoryURL string `json:"repository_url" binding:"required" validate:"required,url" example:"https://github.com/user/repo"`
	// Optional branch name, defaults to main
	Branch string `json:"branch,omitempty" validate:"omitempty,min=1" example:"main"`
	// Optional custom agent name
	AgentName string `json:"agent_name,omitempty" validate:"omitempty,min=1" example:"my-code-analyzer"`
	// Optional AI configuration (defaults to platform default if not provided)
	// AIConfig *config.AIConfiguration `json:"ai_config,omitempty"` // TEMPORARILY DISABLED - will use new AgentAIConfig
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

// UpdateAgentRequest represents the request to update an agent
type UpdateAgentRequest struct {
	// Agent ID (set from URL path)
	AgentID string `json:"-" uri:"agent_id" validate:"required"`
	// Optional new agent name
	AgentName *string `json:"agent_name,omitempty" validate:"omitempty,min=1" example:"updated-code-analyzer"`
	// Optional updated repository URL
	RepositoryURL *string `json:"repository_url,omitempty" validate:"omitempty,url" example:"https://github.com/user/updated-repo"`
	// Optional updated branch name
	Branch *string `json:"branch,omitempty" validate:"omitempty,min=1" example:"develop"`
	// Optional AI configuration updates
	// AIConfig *config.AIConfiguration `json:"ai_config,omitempty"` // TEMPORARILY DISABLED - will use new AgentAIConfig
} //@name UpdateAgentRequest

// UpdateAgentResponse represents the response when updating an agent
type UpdateAgentResponse struct {
	// Unique identifier for the agent
	AgentID string `json:"agent_id" example:"agent-12345"`
	// Agent version
	AgentVersion string `json:"agent_version" example:"v1.1.0"`
	// Knowledge base ID associated with the agent
	KnowledgeBaseID string `json:"knowledge_base_id" example:"kb-67890"`
	// Vector store ID for the knowledge base
	VectorStoreID string `json:"vector_store_id" example:"vs-abcde"`
	// Repository URL
	RepositoryURL string `json:"repository_url" example:"https://github.com/user/updated-repo"`
	// Branch name
	Branch string `json:"branch" example:"develop"`
	// Agent name
	AgentName string `json:"agent_name" example:"updated-code-analyzer"`
	// Agent status
	Status string `json:"status" example:"ready"`
	// Timestamp when the agent was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Timestamp when the agent was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T11:45:00Z"`
} //@name UpdateAgentResponse

// ErrorResponse represents an error response
type ErrorResponse struct {
	// Error code
	Code int `json:"code" example:"400"`
	// Error message
	Message string `json:"message" example:"Invalid request parameters"`
	// Optional error details
	Details string `json:"details,omitempty" example:"repository_url is required"`
} //@name ErrorResponse

// SuccessResponse represents a success response
type SuccessResponse struct {
	// Success message
	Message string `json:"message" example:"Operation completed successfully"`
} //@name SuccessResponse
