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
	factoryMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/factory/mocks"
)

func TestNewAgentService(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)
	mockInfraFactory := factoryMocks.NewMockAIInfrastructureFactory(ctrl)

	// Act
	service := NewDefaultAgentService(mockAgentRepo, mockInfraFactory)

	// Assert
	assert.NotNil(t, service)
	assert.IsType(t, &DefaultAgentService{}, service)
}

func TestDefaultAgentService_GetAgent_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)
	mockInfraFactory := factoryMocks.NewMockAIInfrastructureFactory(ctrl)

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

	service := NewDefaultAgentService(mockAgentRepo, mockInfraFactory)

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

	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)
	mockInfraFactory := factoryMocks.NewMockAIInfrastructureFactory(ctrl)

	agentID := "non-existent-agent"
	expectedError := errors.New("agent not found")

	mockAgentRepo.EXPECT().
		GetAgent(gomock.Any(), agentID).
		Return(nil, expectedError)

	service := NewDefaultAgentService(mockAgentRepo, mockInfraFactory)

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

	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)
	mockInfraFactory := factoryMocks.NewMockAIInfrastructureFactory(ctrl)

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

	service := NewDefaultAgentService(mockAgentRepo, mockInfraFactory)

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

	mockAgentRepo := repoMocks.NewMockAgentRepository(ctrl)
	mockInfraFactory := factoryMocks.NewMockAIInfrastructureFactory(ctrl)

	repoError := errors.New("DynamoDB error")

	mockAgentRepo.EXPECT().
		ListAgents(gomock.Any()).
		Return(nil, repoError).
		Times(1)

	service := NewDefaultAgentService(mockAgentRepo, mockInfraFactory)

	// Act
	request := models.ListAgentsRequest{}
	response, err := service.ListAgents(context.Background(), request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to list agents")
}
