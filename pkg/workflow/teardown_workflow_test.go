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

func TestNewTeardownWorkflow(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Act
	wf, err := workflow.NewTeardownWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestNewTeardownWorkflowWithResources(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	// Act
	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestTeardownWorkflow_Run_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)
}

func TestTeardownWorkflow_Run_AgentTearDownError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	agentError := errors.New("agent teardown failed")

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(agentError).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to tear down agent")
}

func TestTeardownWorkflow_Run_RAGTearDownError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	ragError := errors.New("rag teardown failed")

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(ragError).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to tear down RAG pipeline")
}

func TestTeardownWorkflow_Run_RepositoryCleanupError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	repoError := errors.New("repository cleanup failed")

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(repoError).
		Times(1)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clean up repository")
}

func TestTeardownWorkflow_Run_AllErrors(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	agentError := errors.New("agent teardown failed")
	ragError := errors.New("rag teardown failed")
	repoError := errors.New("repository cleanup failed")

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(agentError).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(ragError).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(repoError).
		Times(1)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.Error(t, err)
	// Should return the first error (agent teardown)
	assert.Contains(t, err.Error(), "failed to tear down agent")
}

func TestTeardownWorkflow_Run_EmptyResourceIDs(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Set up expectations - only repo cleanup should be called
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	// Agent and RAG builders should not be called since resource IDs are empty
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	wf, err := workflow.NewTeardownWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)
}

func TestTeardownWorkflow_SetResourceIDs(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	wf, err := workflow.NewTeardownWorkflow(mockRepo, mockRAGBuilder, mockAgentBuilder)
	require.NoError(t, err)

	// Cast to concrete type to access SetResourceIDs method
	teardownWf := wf.(*workflow.TeardownWorkflow)

	// Act
	teardownWf.SetResourceIDs(vectorStoreID, ragID, agentID, agentVersion)

	// Set up expectations for cleanup with the set resource IDs
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(nil).
		Times(1)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	// Assert
	err = wf.Run(context.Background())
	assert.NoError(t, err)
}

func TestTeardownWorkflow_Run_PartialResourceIDs(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	// Only set agent ID, not RAG ID
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	// Set up expectations - only repo cleanup should be called
	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		Times(1)

	// Agent and RAG builders should not be called since we don't have both required IDs
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		"", "", agentID, agentVersion, // Empty vectorStoreID and ragID
	)
	require.NoError(t, err)

	// Act
	err = wf.Run(context.Background())

	// Assert
	assert.NoError(t, err)
}

func BenchmarkTeardownWorkflow_Run(b *testing.B) {
	// Setup
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockRepository(ctrl)
	mockRAGBuilder := builderMocks.NewMockRAGBuilder(ctrl)
	mockAgentBuilder := builderMocks.NewMockAgentBuilder(ctrl)

	vectorStoreID := "test-vector-store"
	ragID := "test-rag-id"
	agentID := "test-agent-id"
	agentVersion := "test-agent-version"

	// Set up expectations
	mockAgentBuilder.EXPECT().
		TearDown(gomock.Any(), agentID, agentVersion, ragID).
		Return(nil).
		AnyTimes()

	mockRAGBuilder.EXPECT().
		TearDown(gomock.Any(), vectorStoreID, ragID).
		Return(nil).
		AnyTimes()

	mockRepo.EXPECT().
		Cleanup().
		Return(nil).
		AnyTimes()

	wf, err := workflow.NewTeardownWorkflowWithResources(
		mockRepo, mockRAGBuilder, mockAgentBuilder,
		vectorStoreID, ragID, agentID, agentVersion,
	)
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
