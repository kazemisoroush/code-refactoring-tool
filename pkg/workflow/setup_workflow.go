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

// SetupWorkflow represents a workflow for setting up AI resources
type SetupWorkflow struct {
	Config       config.Config
	Repository   repository.Repository
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs created during setup
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewSetupWorkflow creates a new SetupWorkflow instance
func NewSetupWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &SetupWorkflow{
		Config:       cfg,
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// Run executes the setup workflow to provision AI resources
func (s *SetupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running setup workflow")

	// 1. Clone the repository
	slog.Info("Cloning repository")
	err := s.Repository.Cleanup()
	if err != nil {
		return fmt.Errorf("failed to clean up repository: %w", err)
	}
	slog.Info("Repository cleaned up successfully")

	err = s.Repository.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	slog.Info("Repository cloned successfully")

	// 2. Build the RAG pipeline
	slog.Info("Building RAG pipeline")
	ragID, err := s.RAGBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build RAG pipeline: %w", err)
	}
	s.RAGID = ragID
	s.VectorStoreID = ragID // In Bedrock, the KB ID serves as both RAG ID and vector store ID
	slog.Info("RAG pipeline built successfully", "ragID", ragID)

	// 3. Build the agent
	slog.Info("Building agent", "ragID", ragID)
	agentID, agentVersion, err := s.AgentBuilder.Build(ctx, ragID)
	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}
	s.AgentID = agentID
	s.AgentVersion = agentVersion
	slog.Info("Agent built successfully", "agentID", agentID, "version", agentVersion)

	slog.Info("Setup workflow completed successfully")
	return nil
}

// GetResourceIDs returns the resource IDs created during setup
func (s *SetupWorkflow) GetResourceIDs() (vectorStoreID, ragID, agentID, agentVersion string) {
	return s.VectorStoreID, s.RAGID, s.AgentID, s.AgentVersion
}
