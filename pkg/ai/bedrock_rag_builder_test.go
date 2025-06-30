package ai_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai"
	mocks_rag "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag/mocks"
	mocks_storage "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage/mocks"
	mocks_vector "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/vector/mocks"
	mocks_repo "github.com/kazemisoroush/code-refactoring-tool/pkg/repository/mocks"
	"github.com/stretchr/testify/require"
)

func TestBedrockRAGBuilder_Build(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks_repo.NewMockRepository(ctrl)
	repo.EXPECT().GetPath().Return("test-repo-path").AnyTimes()

	storage := mocks_storage.NewMockStorage(ctrl)
	storage.EXPECT().UploadDirectory(ctx, "test-repo-path", "test-repo-path").Return(nil).AnyTimes()

	vectorDataStore := mocks_vector.NewMockVectorStore(ctrl)
	vectorDataStore.EXPECT().EnsureSchema(ctx, gomock.Any()).Return(nil).AnyTimes()

	rag := mocks_rag.NewMockRAG(ctrl)
	rag.EXPECT().Create(ctx, gomock.Any()).Return("test-kb-id", nil).AnyTimes()

	builder := ai.NewBedrockRAGBuilder(
		repo,
		storage,
		vectorDataStore,
		rag,
	)

	// Act
	ragMetadata, err := builder.Build(ctx)

	// Assert
	require.NoError(t, err, "Failed to build RAG pipeline")
	require.NotNil(t, ragMetadata, "RAG metadata should not be nil")
}
