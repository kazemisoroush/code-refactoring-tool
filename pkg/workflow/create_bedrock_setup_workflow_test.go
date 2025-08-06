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

func TestNewCreateBedrockSetupWorkflow(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Act
	wf, err := workflow.NewCreateBedrockSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestCreateBedrockSetupWorkflow_Run_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "bedrock-rag-id"
	agentID := "bedrock-agent-id"
	agentVersion := "bedrock-agent-version"

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

	wf, err := workflow.NewCreateBedrockSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)

	// Cast to concrete type to access GetResourceIDs method
	setupWf := wf.(*workflow.CreateBedrockSetupWorkflow)
	vectorStoreID, returnedRAGID, returnedAgentID, returnedAgentVersion := setupWf.GetResourceIDs()

	assert.Equal(t, ragID, vectorStoreID) // VectorStoreID should be same as RAG ID for Bedrock
	assert.Equal(t, ragID, returnedRAGID)
	assert.Equal(t, agentID, returnedAgentID)
	assert.Equal(t, agentVersion, returnedAgentVersion)
}

func TestCreateBedrockSetupWorkflow_Run_RAGBuildError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragError := errors.New("bedrock rag build failed")

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
		Return("", ragError).
		Times(1)

	// Agent builder should not be called since RAG build fails
	mockAgentBuilder.EXPECT().
		Build(gomock.Any(), gomock.Any()).
		Times(0)

	wf, err := workflow.NewCreateBedrockSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build Bedrock RAG pipeline")
}
