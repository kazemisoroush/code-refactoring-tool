package builder_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	mocks_rag "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag/mocks"
	mocks_storage "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage/mocks"
	"github.com/stretchr/testify/require"
)

func TestBedrockRAGBuilder_Build(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoPath := "test-repo-path"

	dataStore := mocks_storage.NewMockDataStore(ctrl)
	dataStore.EXPECT().UploadDirectory(ctx, "test-repo-path", "test-repo-path").Return(nil).Times(1)
	dataStore.EXPECT().Create(ctx, gomock.Any()).Return("test-data-source-id", nil).Times(1)

	storage := mocks_storage.NewMockStorage(ctrl)
	storage.EXPECT().EnsureSchema(ctx, "code_refactoring_db", "test-repo-path").Return(nil).Times(1)

	rag := mocks_rag.NewMockRAG(ctrl)
	rag.EXPECT().Create(ctx, gomock.Any()).Return("test-kb-id", nil).Times(1)

	builder := builder.NewBedrockRAGBuilder(
		repoPath,
		dataStore,
		storage,
		rag,
	)

	// Act
	ragMetadata, err := builder.Build(ctx)

	// Assert
	require.NoError(t, err, "Failed to build RAG pipeline")
	require.NotNil(t, ragMetadata, "RAG metadata should not be nil")
}
