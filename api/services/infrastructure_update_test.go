package services

import (
	"testing"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
)

func TestRequiresInfrastructureUpdate(t *testing.T) {
	// Helper function to create a pointer to a string
	stringPtr := func(s string) *string {
		return &s
	}

	// Helper function to create a pointer to an AIProvider
	aiProviderPtr := func(p models.AIProvider) *models.AIProvider {
		return &p
	}

	// Create a base existing agent record for testing
	baseExistingAgent := &repository.AgentRecord{
		AgentID:       "agent-123",
		AgentName:     "test-agent",
		RepositoryURL: "https://github.com/user/repo",
		Branch:        "main",
		AIProvider:    "bedrock",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	tests := []struct {
		name           string
		request        models.UpdateAgentRequest
		existingAgent  *repository.AgentRecord
		expectedResult bool
		description    string
	}{
		{
			name: "No changes - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "When no fields are updated, infrastructure should remain unchanged",
		},
		{
			name: "Only agent name changed - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:   "agent-123",
				AgentName: stringPtr("new-agent-name"),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "Agent name changes don't affect infrastructure",
		},
		{
			name: "Repository URL changed - should require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr("https://github.com/user/new-repo"),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: true,
			description:    "Repository URL changes require infrastructure update to reindex the codebase",
		},
		{
			name: "Branch changed - should require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
				Branch:  stringPtr("develop"),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: true,
			description:    "Branch changes require infrastructure update to reindex the new branch",
		},
		{
			name: "AI Provider changed - should require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:    "agent-123",
				AIProvider: aiProviderPtr(models.AIProviderOpenAI),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: true,
			description:    "AI provider changes require infrastructure update to recreate resources with new provider",
		},
		{
			name: "Multiple non-infrastructure fields changed - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:   "agent-123",
				AgentName: stringPtr("updated-name"),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "Only metadata changes don't require infrastructure updates",
		},
		{
			name: "Repository URL same as existing - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr("https://github.com/user/repo"), // Same as existing
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "Setting the same repository URL shouldn't trigger infrastructure update",
		},
		{
			name: "Branch same as existing - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
				Branch:  stringPtr("main"), // Same as existing
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "Setting the same branch shouldn't trigger infrastructure update",
		},
		{
			name: "AI Provider same as existing - should not require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:    "agent-123",
				AIProvider: aiProviderPtr(models.AIProviderBedrock), // Same as existing
			},
			existingAgent:  baseExistingAgent,
			expectedResult: false,
			description:    "Setting the same AI provider shouldn't trigger infrastructure update",
		},
		{
			name: "Multiple infrastructure fields changed - should require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr("https://github.com/user/different-repo"),
				Branch:        stringPtr("feature-branch"),
				AIProvider:    aiProviderPtr(models.AIProviderLocal),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: true,
			description:    "Multiple infrastructure changes should still return true",
		},
		{
			name: "Mixed infrastructure and non-infrastructure changes - should require infrastructure update",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				AgentName:     stringPtr("new-name"),
				RepositoryURL: stringPtr("https://github.com/user/different-repo"),
			},
			existingAgent:  baseExistingAgent,
			expectedResult: true,
			description:    "Any infrastructure change should trigger update regardless of other changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresInfrastructureUpdate(tt.request, tt.existingAgent)

			if result != tt.expectedResult {
				t.Errorf("RequiresInfrastructureUpdate() = %v, expected %v", result, tt.expectedResult)
				t.Errorf("Description: %s", tt.description)
				t.Errorf("Request: %+v", tt.request)
				t.Errorf("Existing Agent: %+v", tt.existingAgent)
			}
		})
	}
}

// TestRequiresInfrastructureUpdate_EdgeCases tests edge cases and boundary conditions
func TestRequiresInfrastructureUpdate_EdgeCases(t *testing.T) {
	stringPtr := func(s string) *string {
		return &s
	}

	tests := []struct {
		name           string
		request        models.UpdateAgentRequest
		existingAgent  *repository.AgentRecord
		expectedResult bool
		description    string
	}{
		{
			name: "Empty strings vs nil pointers",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr(""), // Empty string
			},
			existingAgent: &repository.AgentRecord{
				AgentID:       "agent-123",
				RepositoryURL: "https://github.com/user/repo", // Non-empty
			},
			expectedResult: true,
			description:    "Empty string should be considered different from non-empty string",
		},
		{
			name: "Empty string to empty string",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
				Branch:  stringPtr(""), // Empty string
			},
			existingAgent: &repository.AgentRecord{
				AgentID: "agent-123",
				Branch:  "", // Empty string
			},
			expectedResult: false,
			description:    "Empty string to empty string should not require update",
		},
		{
			name: "Whitespace differences",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
				Branch:  stringPtr(" main "), // With whitespace
			},
			existingAgent: &repository.AgentRecord{
				AgentID: "agent-123",
				Branch:  "main", // Without whitespace
			},
			expectedResult: true,
			description:    "Whitespace differences should be considered as changes",
		},
		{
			name: "Case sensitivity",
			request: models.UpdateAgentRequest{
				AgentID: "agent-123",
				Branch:  stringPtr("Main"), // Capital M
			},
			existingAgent: &repository.AgentRecord{
				AgentID: "agent-123",
				Branch:  "main", // Lowercase m
			},
			expectedResult: true,
			description:    "Case differences should be considered as changes",
		},
		{
			name: "URL with different protocols",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr("http://github.com/user/repo"), // http
			},
			existingAgent: &repository.AgentRecord{
				AgentID:       "agent-123",
				RepositoryURL: "https://github.com/user/repo", // https
			},
			expectedResult: true,
			description:    "Protocol differences in URLs should be considered as changes",
		},
		{
			name: "URL with trailing slash",
			request: models.UpdateAgentRequest{
				AgentID:       "agent-123",
				RepositoryURL: stringPtr("https://github.com/user/repo/"), // With trailing slash
			},
			existingAgent: &repository.AgentRecord{
				AgentID:       "agent-123",
				RepositoryURL: "https://github.com/user/repo", // Without trailing slash
			},
			expectedResult: true,
			description:    "Trailing slash differences should be considered as changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresInfrastructureUpdate(tt.request, tt.existingAgent)

			if result != tt.expectedResult {
				t.Errorf("RequiresInfrastructureUpdate() = %v, expected %v", result, tt.expectedResult)
				t.Errorf("Description: %s", tt.description)
			}
		})
	}
}

// BenchmarkRequiresInfrastructureUpdate benchmarks the function performance
func BenchmarkRequiresInfrastructureUpdate(b *testing.B) {
	stringPtr := func(s string) *string {
		return &s
	}

	request := models.UpdateAgentRequest{
		AgentID:       "agent-123",
		AgentName:     stringPtr("new-name"),
		RepositoryURL: stringPtr("https://github.com/user/new-repo"),
		Branch:        stringPtr("develop"),
	}

	existingAgent := &repository.AgentRecord{
		AgentID:       "agent-123",
		AgentName:     "old-name",
		RepositoryURL: "https://github.com/user/old-repo",
		Branch:        "main",
		AIProvider:    "bedrock",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RequiresInfrastructureUpdate(request, existingAgent)
	}
}
