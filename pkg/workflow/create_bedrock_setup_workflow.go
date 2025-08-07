// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
)

// CreateBedrockSetupWorkflow represents a workflow for setting up Bedrock AI resources
type CreateBedrockSetupWorkflow struct {
	repository   codebase.Codebase
	ragBuilder   builder.RAGBuilder
	agentBuilder builder.AgentBuilder

	// Resource IDs created during setup
	vectorStoreID string
	ragID         string
	agentID       string
	agentVersion  string
}

// NewCreateBedrockSetupWorkflow creates a new CreateBedrockSetupWorkflow instance
func NewCreateBedrockSetupWorkflow(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &CreateBedrockSetupWorkflow{
		repository:   repo,
		ragBuilder:   ragBuilder,
		agentBuilder: agentBuilder,
	}, nil
}

// Run executes the Bedrock setup workflow to provision AI resources
func (s *CreateBedrockSetupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running Bedrock setup workflow")

	defer func() {
		err := s.repository.Cleanup()
		if err != nil {
			slog.Error("failed to cleanup repository", "error", err)
		}
	}()

	// 1. Clone the repository
	slog.Info("Cloning repository for Bedrock setup")
	err := s.repository.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	slog.Info("Repository cloned successfully")

	// 2. Build the RAG pipeline using Bedrock
	slog.Info("Building Bedrock RAG pipeline")
	ragID, err := s.ragBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build Bedrock RAG pipeline: %w", err)
	}
	s.ragID = ragID
	s.vectorStoreID = ragID // In Bedrock, the KB ID serves as both RAG ID and vector store ID
	slog.Info("Bedrock RAG pipeline built successfully", "ragID", ragID)

	// 3. Build the Bedrock agent
	slog.Info("Building Bedrock agent", "ragID", ragID)
	agentID, agentVersion, err := s.agentBuilder.Build(ctx, ragID)
	if err != nil {
		return fmt.Errorf("failed to build Bedrock agent: %w", err)
	}
	s.agentID = agentID
	s.agentVersion = agentVersion
	slog.Info("Bedrock agent built successfully", "agentID", agentID, "version", agentVersion)

	slog.Info("Bedrock setup workflow completed successfully")
	return nil
}

// GetResourceIDs returns the resource IDs created during Bedrock setup
func (s *CreateBedrockSetupWorkflow) GetResourceIDs() (vectorStoreID, ragID, agentID, agentVersion string) {
	return s.vectorStoreID, s.ragID, s.agentID, s.agentVersion
}
