// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalRAGBuilder implements the RAGBuilder interface for local development.
type LocalRAGBuilder struct {
	repoPath       string
	chromaURL      string
	embeddingModel string
}

// NewLocalRAGBuilder creates a new instance of LocalRAGBuilder.
func NewLocalRAGBuilder(repoPath, chromaURL, embeddingModel string) RAGBuilder {
	return &LocalRAGBuilder{
		repoPath:       repoPath,
		chromaURL:      chromaURL,
		embeddingModel: embeddingModel,
	}
}

// Build implements the RAGBuilder interface by creating a local RAG pipeline.
func (l *LocalRAGBuilder) Build(ctx context.Context) (string, error) {
	// Generate a unique RAG ID
	ragID := uuid.New().String()

	// In a real implementation, you would:
	// 1. Scan the repository for relevant files
	// 2. Extract and chunk the code content
	// 3. Generate embeddings using a local model
	// 4. Store embeddings in ChromaDB

	fmt.Printf("Building local RAG pipeline for repository: %s\n", l.repoPath)

	// Simulate scanning the repository
	err := l.scanRepository(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to scan repository: %w", err)
	}

	fmt.Printf("Local RAG pipeline created with ID: %s\n", ragID)

	return ragID, nil
}

// TearDown implements the RAGBuilder interface by cleaning up local RAG resources.
func (l *LocalRAGBuilder) TearDown(_ context.Context, vectorStoreID string, ragID string) error {
	// In a real implementation, you would:
	// 1. Delete the ChromaDB collection
	// 2. Clean up any temporary files
	// 3. Remove cached embeddings

	fmt.Printf("Local RAG pipeline torn down - vector store: %s, RAG ID: %s\n", vectorStoreID, ragID)

	return nil
}

// scanRepository simulates scanning the repository for code files.
func (l *LocalRAGBuilder) scanRepository(_ context.Context) error {
	var fileCount int

	err := filepath.Walk(l.repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-code files
		if info.IsDir() {
			return nil
		}

		// Only process certain file types for code analysis
		ext := strings.ToLower(filepath.Ext(path))
		codeExtensions := map[string]bool{
			".go":    true,
			".js":    true,
			".ts":    true,
			".py":    true,
			".java":  true,
			".cpp":   true,
			".c":     true,
			".h":     true,
			".hpp":   true,
			".cs":    true,
			".rb":    true,
			".php":   true,
			".swift": true,
			".kt":    true,
			".rs":    true,
		}

		if codeExtensions[ext] {
			fileCount++
			// In a real implementation, you would process the file content here
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("Scanned %d code files in repository\n", fileCount)
	return nil
}
