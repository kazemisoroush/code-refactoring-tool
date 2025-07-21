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
	repositoryMocks "github.com/kazemisoroush/code-refactoring-tool/pkg/repository/mocks"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

func TestNewSetupWorkflow(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Act
	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestSetupWorkflow_Run_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

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

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)

	// Cast to concrete type to access GetResourceIDs method
	setupWf := wf.(*workflow.SetupWorkflow)
	vectorStoreID, returnedRAGID, returnedAgentID, returnedAgentVersion := setupWf.GetResourceIDs()

	assert.Equal(t, ragID, vectorStoreID) // VectorStoreID should be same as RAG ID
	assert.Equal(t, ragID, returnedRAGID)
	assert.Equal(t, agentID, returnedAgentID)
	assert.Equal(t, agentVersion, returnedAgentVersion)
}

func TestSetupWorkflow_Run_RepositoryCloneError(t *testing.T) {
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

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone repository")
}

func TestSetupWorkflow_Run_RAGBuildError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragError := errors.New("rag build failed")

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

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build RAG pipeline")
}

func TestSetupWorkflow_Run_AgentBuildError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "test-rag-id"
	agentError := errors.New("agent build failed")

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

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build agent")
}

func TestSetupWorkflow_GetResourceIDs(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Cast to concrete type to access GetResourceIDs method
	setupWf := wf.(*workflow.SetupWorkflow)

	// Act - get initial resource IDs (should be empty)
	vectorStoreID, ragID, agentID, agentVersion := setupWf.GetResourceIDs()

	// Assert
	assert.Equal(t, "", vectorStoreID)
	assert.Equal(t, "", ragID)
	assert.Equal(t, "", agentID)
	assert.Equal(t, "", agentVersion)
}

func BenchmarkSetupWorkflow_Run(b *testing.B) {
	// Setup
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	// Set up expectations
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		AnyTimes()

	mockRepo.EXPECT().
		Clone(gomock.Any()).
		Return(nil).
		AnyTimes()

	mockRAGBuilder.EXPECT().
		Build(gomock.Any()).
		Return(ragID, nil).
		AnyTimes()

	mockAgentBuilder.EXPECT().
		Build(gomock.Any(), ragID).
		Return(agentID, agentVersion, nil).
		AnyTimes()

	wf, err := workflow.NewSetupWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	if err != nil {
		b.Fatal(err)
	}

	// Reset timer before running the benchmark
	b.ResetTimer()

	// Run benchmark
	for i := 0; i < b.N; i++ {
		if err := wf.Run(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
