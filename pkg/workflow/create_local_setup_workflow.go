// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
)

// CreateLocalSetupWorkflow represents a workflow for setting up local AI resources
type CreateLocalSetupWorkflow struct {
	Repository   codebase.Codebase
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs created during setup
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewCreateLocalSetupWorkflow creates a new CreateLocalSetupWorkflow instance
func NewCreateLocalSetupWorkflow(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &CreateLocalSetupWorkflow{
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// Run executes the local setup workflow to provision AI resources
func (s *CreateLocalSetupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running local setup workflow")

	defer func() {
		err := s.Repository.Cleanup()
		if err != nil {
			slog.Error("failed to cleanup repository", "error", err)
		}
	}()

	// 1. Clone the repository
	slog.Info("Cloning repository for local setup")
	err := s.Repository.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	slog.Info("Repository cloned successfully")

	// 2. Build the local RAG pipeline
	slog.Info("Building local RAG pipeline")
	ragID, err := s.RAGBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build local RAG pipeline: %w", err)
	}
	s.RAGID = ragID
	s.VectorStoreID = ragID // For local setup, use the same ID for both
	slog.Info("Local RAG pipeline built successfully", "ragID", ragID)

	// 3. Build the local agent
	slog.Info("Building local agent", "ragID", ragID)
	agentID, agentVersion, err := s.AgentBuilder.Build(ctx, ragID)
	if err != nil {
		return fmt.Errorf("failed to build local agent: %w", err)
	}
	s.AgentID = agentID
	s.AgentVersion = agentVersion
	slog.Info("Local agent built successfully", "agentID", agentID, "version", agentVersion)

	slog.Info("Local setup workflow completed successfully")
	return nil
}

// GetResourceIDs returns the resource IDs created during local setup
func (s *CreateLocalSetupWorkflow) GetResourceIDs() (vectorStoreID, ragID, agentID, agentVersion string) {
	return s.VectorStoreID, s.RAGID, s.AgentID, s.AgentVersion
}
