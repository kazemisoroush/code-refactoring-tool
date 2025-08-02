// Package models contains data structures for AI configuration and provider management
package models

import (
	"time"
)

// AIProvider represents the available AI infrastructure providers
type AIProvider string

const (
	// AIProviderBedrock uses AWS Bedrock (default for SaaS)
	AIProviderBedrock AIProvider = "bedrock"
	// AIProviderLocal uses local Ollama + ChromaDB (for development/enterprise)
	AIProviderLocal AIProvider = "local"
	// AIProviderOpenAI uses OpenAI APIs (future extension)
	AIProviderOpenAI AIProvider = "openai"
)

// AIConfiguration represents AI provider-specific configuration
type AIConfiguration struct {
	Provider AIProvider `json:"provider" example:"bedrock"`
	// Local AI configuration (only used if provider is "local")
	Local *LocalAIConfig `json:"local,omitempty"`
	// Bedrock configuration (only used if provider is "bedrock")
	Bedrock *BedrockAIConfig `json:"bedrock,omitempty"`
}

// LocalAIConfig represents configuration for local AI setup
type LocalAIConfig struct {
	OllamaURL      string `json:"ollama_url" example:"http://localhost:11434"`
	Model          string `json:"model" example:"codellama:7b-instruct"`
	ChromaURL      string `json:"chroma_url" example:"http://localhost:8000"`
	EmbeddingModel string `json:"embedding_model" example:"all-MiniLM-L6-v2"`
}

// BedrockAIConfig represents configuration for AWS Bedrock
type BedrockAIConfig struct {
	Region                      string `json:"region,omitempty" example:"us-west-2"`
	KnowledgeBaseServiceRoleARN string `json:"knowledge_base_service_role_arn,omitempty"`
	AgentServiceRoleARN         string `json:"agent_service_role_arn,omitempty"`
	FoundationModel             string `json:"foundation_model,omitempty" example:"amazon.titan-tg1-large"`
}

// AgentConfigurationResponse includes AI provider information
type AgentConfigurationResponse struct {
	AgentID         string     `json:"agent_id" example:"agent-12345"`
	AgentVersion    string     `json:"agent_version" example:"v1.0.0"`
	KnowledgeBaseID string     `json:"knowledge_base_id" example:"kb-67890"`
	VectorStoreID   string     `json:"vector_store_id" example:"vs-abcde"`
	Status          string     `json:"status" example:"created"`
	CreatedAt       time.Time  `json:"created_at"`
	AIProvider      AIProvider `json:"ai_provider" example:"bedrock"`
	// Additional metadata about the AI setup
	AIMetadata map[string]interface{} `json:"ai_metadata,omitempty"`
}
