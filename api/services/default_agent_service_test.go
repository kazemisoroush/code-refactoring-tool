package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	repoMocks "github.com/kazemisoroush/code-refactoring-tool/api/repository/mocks"
	builderMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder/mocks"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	gitRepoMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/repository/mocks"
)

func TestNewAgentService(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gitConfig := config.GitConfig{
		RepoURL: "https://github.com/example/repo.git",
		Token:   "test-token",
		Author:  "Test Author",
		Email:   "test@example.com",
	}
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)
	mockGitRepo := gitRepoMocks.NewMockRepository(ctrl)
	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)

	// Act
	service := NewAgentService(gitConfig, mockRAGBuilder, mockAgentBuilder, mockGitRepo, mockAgentRepo)

	// Assert
	assert.NotNil(t, service)
	assert.IsType(t, &DefaultAgentService{}, service)
}

func TestDefaultAgentService_GetAgent_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gitConfig := config.GitConfig{}
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)
	mockGitRepo := gitRepoMocks.NewMockRepository(ctrl)
	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)

	agentID := "test-agent-id"
	expectedRecord := &repository.AgentRecord{
		AgentID:         agentID,
		AgentVersion:    "v1.0",
		KnowledgeBaseID: "kb-123",
		VectorStoreID:   "vs-123",
		RepositoryURL:   "https://github.com/example/repo",
		Status:          string(models.AgentStatusReady),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockAgentRepo.EXPECT().
		GetAgent(gomock.Any(), agentID).
		Return(expectedRecord, nil)

	service := NewAgentService(gitConfig, mockRAGBuilder, mockAgentBuilder, mockGitRepo, mockAgentRepo)

	// Act
	result, err := service.GetAgent(context.Background(), agentID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, agentID, result.AgentID)
	assert.Equal(t, expectedRecord.AgentVersion, result.AgentVersion)
}

func TestDefaultAgentService_GetAgent_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gitConfig := config.GitConfig{}
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)
	mockGitRepo := gitRepoMocks.NewMockRepository(ctrl)
	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)

	agentID := "non-existent-agent"
	expectedError := errors.New("agent not found")

	mockAgentRepo.EXPECT().
		GetAgent(gomock.Any(), agentID).
		Return(nil, expectedError)

	service := NewAgentService(gitConfig, mockRAGBuilder, mockAgentBuilder, mockGitRepo, mockAgentRepo)

	// Act
	response, err := service.GetAgent(context.Background(), agentID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get agent")
}

func TestDefaultAgentService_ListAgents_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)
	mockGitRepo := gitRepoMocks.NewMockRepository(ctrl)
	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)

	agentRecords := []*repository.AgentRecord{
		{
			AgentID:         "agent-123",
			AgentVersion:    "v1.0.0",
			KnowledgeBaseID: "kb-456",
			VectorStoreID:   "vs-789",
			RepositoryURL:   "https://github.com/user/repo1",
			Status:          string(models.AgentStatusReady),
			CreatedAt:       time.Now().UTC(),
		},
		{
			AgentID:         "agent-456",
			AgentVersion:    "v1.0.0",
			KnowledgeBaseID: "kb-789",
			VectorStoreID:   "vs-abc",
			RepositoryURL:   "https://github.com/user/repo2",
			Status:          string(models.AgentStatusReady),
			CreatedAt:       time.Now().UTC(),
		},
	}

	mockAgentRepo.EXPECT().
		ListAgents(gomock.Any()).
		Return(agentRecords, nil).
		Times(1)

	gitConfig := config.GitConfig{}
	service := NewAgentService(gitConfig, mockRAGBuilder, mockAgentBuilder, mockGitRepo, mockAgentRepo)

	// Act
	request := models.ListAgentsRequest{}
	response, err := service.ListAgents(context.Background(), request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Agents, 2)
	assert.Equal(t, "agent-123", response.Agents[0].AgentID)
	assert.Equal(t, "agent-456", response.Agents[1].AgentID)
}

func TestDefaultAgentService_ListAgents_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)
	mockGitRepo := gitRepoMocks.NewMockRepository(ctrl)
	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)

	repoError := errors.New("DynamoDB error")

	mockAgentRepo.EXPECT().
		ListAgents(gomock.Any()).
		Return(nil, repoError).
		Times(1)

	gitConfig := config.GitConfig{}
	service := NewAgentService(gitConfig, mockRAGBuilder, mockAgentBuilder, mockGitRepo, mockAgentRepo)

	// Act
	request := models.ListAgentsRequest{}
	response, err := service.ListAgents(context.Background(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to list agents")
}
