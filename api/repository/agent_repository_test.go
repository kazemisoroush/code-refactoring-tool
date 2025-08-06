package repository

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRecord_GetAIConfig(t *testing.T) {
	tests := []struct {
		name           string
		aiConfigJSON   string
		expectedConfig *models.AgentAIConfig
		expectedError  bool
		errorContains  string
	}{
		{
			name:           "empty JSON returns nil config",
			aiConfigJSON:   "",
			expectedConfig: nil,
			expectedError:  false,
		},
		{
			name: "valid local AI config",
			aiConfigJSON: `{
				"provider": "local",
				"local": {
					"ollama_url": "http://localhost:11434",
					"model": "llama3.1:latest",
					"chroma_url": "http://localhost:8000",
					"embedding_model": "all-MiniLM-L6-v2",
					"temperature": 0.7,
					"max_tokens": 2048
				}
			}`,
			expectedConfig: &models.AgentAIConfig{
				Provider: models.AIProviderLocal,
				Local: &models.LocalAgentConfig{
					OllamaURL:      "http://localhost:11434",
					Model:          "llama3.1:latest",
					ChromaURL:      "http://localhost:8000",
					EmbeddingModel: "all-MiniLM-L6-v2",
					Temperature:    0.7,
					MaxTokens:      2048,
				},
			},
			expectedError: false,
		},
		{
			name: "valid bedrock AI config",
			aiConfigJSON: `{
				"provider": "bedrock",
				"bedrock": {
					"region": "us-east-1",
					"foundation_model": "anthropic.claude-3-sonnet-20240229-v1:0",
					"embedding_model": "amazon.titan-embed-text-v1",
					"knowledge_base_service_role_arn": "arn:aws:iam::123456789012:role/KBRole",
					"agent_service_role_arn": "arn:aws:iam::123456789012:role/AgentRole",
					"s3_bucket_name": "my-code-bucket",
					"temperature": 0.5,
					"max_tokens": 4096
				}
			}`,
			expectedConfig: &models.AgentAIConfig{
				Provider: models.AIProviderBedrock,
				Bedrock: &models.BedrockAgentConfig{
					Region:                      "us-east-1",
					FoundationModel:             "anthropic.claude-3-sonnet-20240229-v1:0",
					EmbeddingModel:              "amazon.titan-embed-text-v1",
					KnowledgeBaseServiceRoleARN: "arn:aws:iam::123456789012:role/KBRole",
					AgentServiceRoleARN:         "arn:aws:iam::123456789012:role/AgentRole",
					S3BucketName:                "my-code-bucket",
					Temperature:                 0.5,
					MaxTokens:                   4096,
				},
			},
			expectedError: false,
		},
		{
			name: "provider only config",
			aiConfigJSON: `{
				"provider": "bedrock"
			}`,
			expectedConfig: &models.AgentAIConfig{
				Provider: models.AIProviderBedrock,
			},
			expectedError: false,
		},
		{
			name:          "invalid JSON",
			aiConfigJSON:  `{"provider": "local", "invalid": }`,
			expectedError: true,
			errorContains: "failed to unmarshal AI config JSON",
		},
		{
			name:          "malformed JSON",
			aiConfigJSON:  `not valid json at all`,
			expectedError: true,
			errorContains: "failed to unmarshal AI config JSON",
		},
		{
			name:           "empty object",
			aiConfigJSON:   `{}`,
			expectedConfig: &models.AgentAIConfig{},
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create AgentRecord with test data
			record := &AgentRecord{
				AgentID:      "test-agent-123",
				AIConfigJSON: tt.aiConfigJSON,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Call the method under test
			config, err := record.GetAIConfig()

			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				if tt.expectedConfig == nil {
					assert.Nil(t, config)
				} else {
					require.NotNil(t, config)
					assert.Equal(t, tt.expectedConfig.Provider, config.Provider)

					// Verify provider-specific configs
					if tt.expectedConfig.Local != nil {
						require.NotNil(t, config.Local)
						assert.Equal(t, tt.expectedConfig.Local.OllamaURL, config.Local.OllamaURL)
						assert.Equal(t, tt.expectedConfig.Local.Model, config.Local.Model)
						assert.Equal(t, tt.expectedConfig.Local.ChromaURL, config.Local.ChromaURL)
						assert.Equal(t, tt.expectedConfig.Local.EmbeddingModel, config.Local.EmbeddingModel)
						assert.Equal(t, tt.expectedConfig.Local.Temperature, config.Local.Temperature)
						assert.Equal(t, tt.expectedConfig.Local.MaxTokens, config.Local.MaxTokens)
					}

					if tt.expectedConfig.Bedrock != nil {
						require.NotNil(t, config.Bedrock)
						assert.Equal(t, tt.expectedConfig.Bedrock.Region, config.Bedrock.Region)
						assert.Equal(t, tt.expectedConfig.Bedrock.FoundationModel, config.Bedrock.FoundationModel)
						assert.Equal(t, tt.expectedConfig.Bedrock.EmbeddingModel, config.Bedrock.EmbeddingModel)
						assert.Equal(t, tt.expectedConfig.Bedrock.KnowledgeBaseServiceRoleARN, config.Bedrock.KnowledgeBaseServiceRoleARN)
						assert.Equal(t, tt.expectedConfig.Bedrock.AgentServiceRoleARN, config.Bedrock.AgentServiceRoleARN)
						assert.Equal(t, tt.expectedConfig.Bedrock.S3BucketName, config.Bedrock.S3BucketName)
						assert.Equal(t, tt.expectedConfig.Bedrock.Temperature, config.Bedrock.Temperature)
						assert.Equal(t, tt.expectedConfig.Bedrock.MaxTokens, config.Bedrock.MaxTokens)
					}
				}
			}
		})
	}
}

func TestNewAgentRecord_WithAIConfig(t *testing.T) {
	tests := []struct {
		name             string
		request          models.CreateAgentRequest
		expectedProvider string
		hasConfigJSON    bool
	}{
		{
			name: "request with local AI config",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "main",
				AgentName:     "test-agent",
				AIConfig: &models.AgentAIConfig{
					Provider: models.AIProviderLocal,
					Local: &models.LocalAgentConfig{
						OllamaURL: "http://localhost:11434",
						Model:     "llama3.1:latest",
					},
				},
			},
			expectedProvider: "local",
			hasConfigJSON:    true,
		},
		{
			name: "request with bedrock AI config",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "main",
				AgentName:     "test-agent",
				AIConfig: &models.AgentAIConfig{
					Provider: models.AIProviderBedrock,
					Bedrock: &models.BedrockAgentConfig{
						Region:          "us-east-1",
						FoundationModel: "anthropic.claude-3-sonnet-20240229-v1:0",
					},
				},
			},
			expectedProvider: "bedrock",
			hasConfigJSON:    true,
		},
		{
			name: "request without AI config",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "main",
				AgentName:     "test-agent",
			},
			expectedProvider: "",
			hasConfigJSON:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call NewAgentRecord
			record := NewAgentRecord(tt.request, "agent-123", "v1.0.0", "kb-456", "vs-789")

			// Verify basic fields
			assert.Equal(t, "agent-123", record.AgentID)
			assert.Equal(t, "v1.0.0", record.AgentVersion)
			assert.Equal(t, "kb-456", record.KnowledgeBaseID)
			assert.Equal(t, "vs-789", record.VectorStoreID)
			assert.Equal(t, tt.request.RepositoryURL, record.RepositoryURL)
			assert.Equal(t, tt.request.Branch, record.Branch)
			assert.Equal(t, tt.request.AgentName, record.AgentName)

			// Verify AI config handling
			assert.Equal(t, tt.expectedProvider, record.AIProvider)

			if tt.hasConfigJSON {
				assert.NotEmpty(t, record.AIConfigJSON)

				// Verify we can deserialize it back
				config, err := record.GetAIConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.request.AIConfig.Provider, config.Provider)
			} else {
				assert.Empty(t, record.AIConfigJSON)

				config, err := record.GetAIConfig()
				assert.NoError(t, err)
				assert.Nil(t, config)
			}
		})
	}
}

func TestAgentRecord_GetAIConfig_RoundTrip(t *testing.T) {
	// Test that we can serialize and deserialize complex configs without data loss
	originalConfig := &models.AgentAIConfig{
		Provider: models.AIProviderLocal,
		Local: &models.LocalAgentConfig{
			OllamaURL:      "http://localhost:11434",
			Model:          "llama3.1:latest",
			ChromaURL:      "http://localhost:8000",
			EmbeddingModel: "all-MiniLM-L6-v2",
			Temperature:    0.7,
			MaxTokens:      2048,
		},
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(originalConfig)
	require.NoError(t, err)

	// Create record with serialized config
	record := &AgentRecord{
		AgentID:      "test-agent",
		AIConfigJSON: string(configJSON),
	}

	// Deserialize using GetAIConfig
	deserializedConfig, err := record.GetAIConfig()
	require.NoError(t, err)
	require.NotNil(t, deserializedConfig)

	// Verify all fields match
	assert.Equal(t, originalConfig.Provider, deserializedConfig.Provider)
	assert.Equal(t, originalConfig.Local.OllamaURL, deserializedConfig.Local.OllamaURL)
	assert.Equal(t, originalConfig.Local.Model, deserializedConfig.Local.Model)
	assert.Equal(t, originalConfig.Local.ChromaURL, deserializedConfig.Local.ChromaURL)
	assert.Equal(t, originalConfig.Local.EmbeddingModel, deserializedConfig.Local.EmbeddingModel)
	assert.Equal(t, originalConfig.Local.Temperature, deserializedConfig.Local.Temperature)
	assert.Equal(t, originalConfig.Local.MaxTokens, deserializedConfig.Local.MaxTokens)
}
