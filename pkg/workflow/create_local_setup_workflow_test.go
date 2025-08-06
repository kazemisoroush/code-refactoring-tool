// Package workflow_test contains tests for the workflow package.
package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	builderMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder/mocks"
	repositoryMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/codebase/mocks"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

func TestNewCreateLocalSetupWorkflow(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Act
	wf, err := workflow.NewCreateLocalSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestCreateLocalSetupWorkflow_Run_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "local-rag-id"
	agentID := "local-agent-id"
	agentVersion := "local-agent-version"

	// Set up expectations
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Clone(gomock.Any()).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		Build(gomock.Any()).
		Return(ragID, nil).
		Times(1)

	mockAgentBuilder.EXPECT().
		Build(gomock.Any(), ragID).
		Return(agentID, agentVersion, nil).
		Times(1)

	wf, err := workflow.NewCreateLocalSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)

	// Cast to concrete type to access GetResourceIDs method
	setupWf := wf.(*workflow.CreateLocalSetupWorkflow)
	vectorStoreID, returnedRAGID, returnedAgentID, returnedAgentVersion := setupWf.GetResourceIDs()

	assert.Equal(t, ragID, vectorStoreID) // VectorStoreID should be same as RAG ID for local
	assert.Equal(t, ragID, returnedRAGID)
	assert.Equal(t, agentID, returnedAgentID)
	assert.Equal(t, agentVersion, returnedAgentVersion)
}

func TestCreateLocalSetupWorkflow_Run_RepositoryCloneError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	repoError := errors.New("repository clone failed")

	// Set up expectations
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Clone(gomock.Any()).
		Return(repoError).
		Times(1)

	// RAG and Agent builders should not be called since repo clone fails
	mockRAGBuilder.EXPECT().
		Build(gomock.Any()).
		Times(0)

	mockAgentBuilder.EXPECT().
		Build(gomock.Any(), gomock.Any()).
		Times(0)

	wf, err := workflow.NewCreateLocalSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone repository")
}

func TestCreateLocalSetupWorkflow_Run_AgentBuildError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "local-rag-id"
	agentError := errors.New("local agent build failed")

	// Set up expectations
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Clone(gomock.Any()).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		Build(gomock.Any()).
		Return(ragID, nil).
		Times(1)

	mockAgentBuilder.EXPECT().
		Build(gomock.Any(), ragID).
		Return("", "", agentError).
		Times(1)

	wf, err := workflow.NewCreateLocalSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build local agent")
}
