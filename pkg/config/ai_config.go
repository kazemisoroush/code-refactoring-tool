// Package config provides AI configuration structures and types.
package config

// AIProvider represents the available AI infrastructure providers
type AIProvider string

// AIConfiguration represents AI provider-specific configuration
type AIConfiguration struct {
	Provider AIProvider `json:"provider" example:"bedrock"`
	// Local AI configuration (only used if provider is "local")
	Local *LocalAIRequestConfig `json:"local,omitempty"`
	// Bedrock configuration (only used if provider is "bedrock")
	Bedrock *BedrockAIRequestConfig `json:"bedrock,omitempty"`
}

// LocalAIRequestConfig represents configuration for local AI setup (API request level)
type LocalAIRequestConfig struct {
	OllamaURL      string `json:"ollama_url" example:"http://localhost:11434"`
	Model          string `json:"model" example:"codellama:7b-instruct"`
	ChromaURL      string `json:"chroma_url" example:"http://localhost:8000"`
	EmbeddingModel string `json:"embedding_model" example:"all-MiniLM-L6-v2"`
}

// BedrockAIRequestConfig represents configuration for AWS Bedrock (API request level)
type BedrockAIRequestConfig struct {
	Region                      string `json:"region,omitempty" example:"us-west-2"`
	KnowledgeBaseServiceRoleARN string `json:"knowledge_base_service_role_arn,omitempty"`
	AgentServiceRoleARN         string `json:"agent_service_role_arn,omitempty"`
	FoundationModel             string `json:"foundation_model,omitempty" example:"amazon.titan-tg1-large"`
}

// GetProvider returns the AI provider, using constants
func (c *AIConfiguration) GetProvider() string {
	return string(c.Provider)
}

// IsLocal returns true if the configuration specifies local AI provider
func (c *AIConfiguration) IsLocal() bool {
	return c.GetProvider() == AIProviderLocal
}

// IsBedrock returns true if the configuration specifies Bedrock AI provider
func (c *AIConfiguration) IsBedrock() bool {
	return c.GetProvider() == AIProviderBedrock
}

// ValidateProvider ensures the provider is one of the supported options
func (c *AIConfiguration) ValidateProvider() error {
	switch c.GetProvider() {
	case AIProviderBedrock, AIProviderLocal, AIProviderOpenAI:
		return nil
	default:
		return ErrUnsupportedAIProvider
	}
}
