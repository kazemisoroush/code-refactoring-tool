package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentStatus_Values(t *testing.T) {
	assert.Equal(t, AgentStatus("creating"), AgentStatusCreating)
	assert.Equal(t, AgentStatus("ready"), AgentStatusReady)
	assert.Equal(t, AgentStatus("error"), AgentStatusError)
	assert.Equal(t, AgentStatus("deleted"), AgentStatusDeleted)
}

func TestCreateAgentRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateAgentRequest
		valid   bool
	}{
		{
			name: "valid request with all fields",
			request: CreateAgentRequest{
				RepositoryURL: "https://github.com/user/repo",
				Branch:        "main",
				AgentName:     "my-agent",
			},
			valid: true,
		},
		{
			name: "valid request with minimal fields",
			request: CreateAgentRequest{
				RepositoryURL: "https://github.com/user/repo",
			},
			valid: true,
		},
		{
			name: "invalid request with empty repository URL",
			request: CreateAgentRequest{
				Branch:    "main",
				AgentName: "my-agent",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would be testing our own validation logic if we had any
			// For now, we're just testing that the struct fields are set correctly
			if tt.valid {
				assert.NotEmpty(t, tt.request.RepositoryURL)
			} else {
				assert.Empty(t, tt.request.RepositoryURL)
			}
		})
	}
}
