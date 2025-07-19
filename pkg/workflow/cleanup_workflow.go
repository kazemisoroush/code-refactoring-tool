// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// CleanupWorkflow represents a code analysis workflow cleanup
type CleanupWorkflow struct {
	Config       config.Config
	Repository   repository.Repository
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs to clean up
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewCleanupWorkflow creates a new CleanupWorkflow instance
func NewCleanupWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &CleanupWorkflow{
		Config:       cfg,
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// NewCleanupWorkflowWithResources creates a new CleanupWorkflow instance with resource IDs
func NewCleanupWorkflowWithResources(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
	vectorStoreID, ragID, agentID, agentVersion string,
) (Workflow, error) {
	return &CleanupWorkflow{
		Config:        cfg,
		Repository:    repo,
		RAGBuilder:    ragBuilder,
		AgentBuilder:  agentBuilder,
		VectorStoreID: vectorStoreID,
		RAGID:         ragID,
		AgentID:       agentID,
		AgentVersion:  agentVersion,
	}, nil
}

// SetResourceIDs sets the resource IDs for cleanup
func (c *CleanupWorkflow) SetResourceIDs(vectorStoreID, ragID, agentID, agentVersion string) {
	c.VectorStoreID = vectorStoreID
	c.RAGID = ragID
	c.AgentID = agentID
	c.AgentVersion = agentVersion
}

// Run implements Workflow.
func (c *CleanupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running cleanup workflow")

	// Track cleanup errors to report at the end
	var cleanupErrors []error

	// 1. Tear down the agent if we have the IDs
	if c.AgentID != "" && c.RAGID != "" {
		slog.Info("Tearing down agent", "agentID", c.AgentID)
		if err := c.AgentBuilder.TearDown(ctx, c.AgentID, c.AgentVersion, c.RAGID); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to tear down agent: %w", err))
			slog.Error("Failed to tear down agent", "error", err)
		} else {
			slog.Info("Agent torn down successfully")
		}
	}

	// 2. Tear down the RAG pipeline if we have the IDs
	if c.VectorStoreID != "" && c.RAGID != "" {
		slog.Info("Tearing down RAG pipeline", "ragID", c.RAGID)
		if err := c.RAGBuilder.TearDown(ctx, c.VectorStoreID, c.RAGID); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to tear down RAG pipeline: %w", err))
			slog.Error("Failed to tear down RAG pipeline", "error", err)
		} else {
			slog.Info("RAG pipeline torn down successfully")
		}
	}

	// 3. Clean up the repository
	slog.Info("Cleaning up repository")
	if err := c.Repository.Cleanup(); err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to clean up repository: %w", err))
		slog.Error("Failed to clean up repository", "error", err)
	} else {
		slog.Info("Repository cleaned up successfully")
	}

	// Report any cleanup errors
	if len(cleanupErrors) > 0 {
		slog.Error("Cleanup completed with errors", "errorCount", len(cleanupErrors))
		// Return the first error, but log all errors
		for i, err := range cleanupErrors {
			slog.Error("Cleanup error", "index", i+1, "error", err)
		}
		return cleanupErrors[0]
	}

	slog.Info("Cleanup workflow completed successfully")
	return nil
}
