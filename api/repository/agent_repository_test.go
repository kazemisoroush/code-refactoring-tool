package repository

import (
	"testing"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRecord_GetAIProvider(t *testing.T) {
	tests := []struct {
		name             string
		aiProvider       string
		expectedProvider models.AIProvider
	}{
		{
			name:             "empty provider returns empty",
			aiProvider:       "",
			expectedProvider: "",
		},
		{
			name:             "local provider",
			aiProvider:       "local",
			expectedProvider: models.AIProviderLocal,
		},
		{
			name:             "bedrock provider",
			aiProvider:       "bedrock",
			expectedProvider: models.AIProviderBedrock,
		},
		{
			name:             "openai provider",
			aiProvider:       "openai",
			expectedProvider: models.AIProviderOpenAI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := &AgentRecord{
				AIProvider: tt.aiProvider,
			}

			result := record.GetAIProvider()
			assert.Equal(t, tt.expectedProvider, result)
		})
	}
}

func TestNewAgentRecord(t *testing.T) {
	tests := []struct {
		name     string
		request  models.CreateAgentRequest
		expected func(*AgentRecord)
	}{
		{
			name: "create agent record with local provider",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "main",
				AgentName:     "test-agent",
				AIProvider:    models.AIProviderLocal,
			},
			expected: func(record *AgentRecord) {
				assert.Equal(t, "https://github.com/test/repo", record.RepositoryURL)
				assert.Equal(t, "main", record.Branch)
				assert.Equal(t, "test-agent", record.AgentName)
				assert.Equal(t, string(models.AIProviderLocal), record.AIProvider)
			},
		},
		{
			name: "create agent record with bedrock provider",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "develop",
				AgentName:     "bedrock-agent",
				AIProvider:    models.AIProviderBedrock,
			},
			expected: func(record *AgentRecord) {
				assert.Equal(t, "https://github.com/test/repo", record.RepositoryURL)
				assert.Equal(t, "develop", record.Branch)
				assert.Equal(t, "bedrock-agent", record.AgentName)
				assert.Equal(t, string(models.AIProviderBedrock), record.AIProvider)
			},
		},
		{
			name: "create agent record without provider",
			request: models.CreateAgentRequest{
				RepositoryURL: "https://github.com/test/repo",
				Branch:        "main",
				AgentName:     "simple-agent",
			},
			expected: func(record *AgentRecord) {
				assert.Equal(t, "", record.AIProvider)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentID := "test-agent-123"
			agentVersion := "v1.0.0"
			kbID := "kb-456"
			vectorStoreID := "vs-789"

			record := NewAgentRecord(tt.request, agentID, agentVersion, kbID, vectorStoreID)

			require.NotNil(t, record)
			assert.Equal(t, agentID, record.AgentID)
			assert.Equal(t, agentVersion, record.AgentVersion)
			assert.Equal(t, kbID, record.KnowledgeBaseID)
			assert.Equal(t, vectorStoreID, record.VectorStoreID)
			assert.Equal(t, string(models.AgentStatusReady), record.Status)

			tt.expected(record)
		})
	}
}

func TestToResponse(t *testing.T) {
	// Create test AgentRecord
	record := &AgentRecord{
		AgentID:         "agent-123",
		AgentVersion:    "v1.0.0",
		KnowledgeBaseID: "kb-456",
		VectorStoreID:   "vs-789",
		RepositoryURL:   "https://github.com/test/repo",
		Branch:          "main",
		AgentName:       "test-agent",
		Status:          "ready",
		AIProvider:      "local",
		CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Call ToResponse
	response := record.ToResponse()

	// Verify the response
	assert.Equal(t, "agent-123", response.AgentID)
	assert.Equal(t, "v1.0.0", response.AgentVersion)
	assert.Equal(t, "kb-456", response.KnowledgeBaseID)
	assert.Equal(t, "vs-789", response.VectorStoreID)
	assert.Equal(t, "ready", response.Status)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), response.CreatedAt)
}
