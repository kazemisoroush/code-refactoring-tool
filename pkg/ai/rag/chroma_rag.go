package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChromaRAG implements the RAG interface using ChromaDB for local vector storage.
type ChromaRAG struct {
	baseURL    string
	httpClient *http.Client
}

// CreateCollectionRequest represents the request structure for creating a Chroma collection.
type CreateCollectionRequest struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChromaCollection represents a Chroma collection response.
type ChromaCollection struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewChromaRAG creates a new instance of ChromaRAG.
func NewChromaRAG(baseURL string) RAG {
	return &ChromaRAG{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Create implements the RAG interface by creating a new collection in ChromaDB.
func (c *ChromaRAG) Create(ctx context.Context, table string) (string, error) {
	req := CreateCollectionRequest{
		Name: table,
		Metadata: map[string]interface{}{
			"description": "Code refactoring collection",
			"created_at":  time.Now().Unix(),
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/collections", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request to ChromaDB: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("chromaDB returned status %d", resp.StatusCode)
	}

	var collection ChromaCollection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return collection.ID, nil
}

// Delete implements the RAG interface by deleting a collection from ChromaDB.
func (c *ChromaRAG) Delete(ctx context.Context, id string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/api/v1/collections/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send delete request to ChromaDB: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("chromaDB returned status %d for delete", resp.StatusCode)
	}

	return nil
}
