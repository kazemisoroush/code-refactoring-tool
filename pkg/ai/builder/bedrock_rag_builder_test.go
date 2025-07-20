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
	storage.EXPECT().EnsureSchema(ctx, "code_refactoring_db", "test_repo_path").Return(nil).Times(1)

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

func TestBedrockRAGBuilder_TableNameSanitization(t *testing.T) {
	tests := []struct {
		name         string
		repoPath     string
		expectedName string
	}{
		{
			name:         "hyphens replaced with underscores",
			repoPath:     "code-refactoring-test",
			expectedName: "code_refactoring_test",
		},
		{
			name:         "special characters replaced",
			repoPath:     "/path/to/my-repo@v1.0",
			expectedName: "my_repo_v1_0",
		},
		{
			name:         "starts with number gets underscore prefix",
			repoPath:     "123-project",
			expectedName: "_123_project",
		},
		{
			name:         "already valid name stays same",
			repoPath:     "valid_repo_name",
			expectedName: "valid_repo_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks (won't be used in this test)
			dataStore := mocks_storage.NewMockDataStore(ctrl)
			storage := mocks_storage.NewMockStorage(ctrl)
			rag := mocks_rag.NewMockRAG(ctrl)

			builder := builder.NewBedrockRAGBuilder(
				tt.repoPath,
				dataStore,
				storage,
				rag,
			)

			// Use reflection to access the private method for testing
			// Since the method is private, we'll test it indirectly by checking
			// that EnsureSchema is called with the expected sanitized name
			ctx := context.Background()

			storage.EXPECT().EnsureSchema(ctx, "code_refactoring_db", tt.expectedName).Return(nil).Times(1)
			dataStore.EXPECT().UploadDirectory(ctx, tt.repoPath, tt.repoPath).Return(nil).Times(1)
			dataStore.EXPECT().Create(ctx, gomock.Any()).Return("test-data-source-id", nil).Times(1)
			rag.EXPECT().Create(ctx, tt.expectedName).Return("test-kb-id", nil).Times(1)

			_, err := builder.Build(ctx)
			require.NoError(t, err)
		})
	}
}
