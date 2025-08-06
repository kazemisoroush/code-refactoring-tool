// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
)

// TeardownBedrockSetupWorkflow represents a workflow for tearing down Bedrock AI resources
type TeardownBedrockSetupWorkflow struct {
	Repository   codebase.Codebase
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs to tear down
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewTeardownBedrockSetupWorkflow creates a new TeardownBedrockSetupWorkflow instance
func NewTeardownBedrockSetupWorkflow(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &TeardownBedrockSetupWorkflow{
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// NewTeardownBedrockSetupWorkflowWithResources creates a new TeardownBedrockSetupWorkflow instance with resource IDs
func NewTeardownBedrockSetupWorkflowWithResources(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
	vectorStoreID, ragID, agentID, agentVersion string,
) (Workflow, error) {
	return &TeardownBedrockSetupWorkflow{
		Repository:    repo,
		RAGBuilder:    ragBuilder,
		AgentBuilder:  agentBuilder,
		VectorStoreID: vectorStoreID,
		RAGID:         ragID,
		AgentID:       agentID,
		AgentVersion:  agentVersion,
	}, nil
}

// SetResourceIDs sets the resource IDs for Bedrock teardown
func (t *TeardownBedrockSetupWorkflow) SetResourceIDs(vectorStoreID, ragID, agentID, agentVersion string) {
	t.VectorStoreID = vectorStoreID
	t.RAGID = ragID
	t.AgentID = agentID
	t.AgentVersion = agentVersion
}

// Run implements Workflow for tearing down Bedrock resources.
func (t *TeardownBedrockSetupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running Bedrock teardown workflow")

	defer func() {
		err := t.Repository.Cleanup()
		if err != nil {
			slog.Error("failed to cleanup repository", "error", err)
		}
	}()

	// Track teardown errors to report at the end
	var teardownErrors []error

	// 1. Tear down the Bedrock agent if we have the IDs
	if t.AgentID != "" && t.RAGID != "" {
		slog.Info("Tearing down Bedrock agent", "agentID", t.AgentID)
		if err := t.AgentBuilder.TearDown(ctx, t.AgentID, t.AgentVersion, t.RAGID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down Bedrock agent: %w", err))
			slog.Error("Failed to tear down Bedrock agent", "error", err)
		} else {
			slog.Info("Bedrock agent torn down successfully")
		}
	}

	// 2. Tear down the Bedrock RAG pipeline if we have the IDs
	if t.VectorStoreID != "" && t.RAGID != "" {
		slog.Info("Tearing down Bedrock RAG pipeline", "ragID", t.RAGID)
		if err := t.RAGBuilder.TearDown(ctx, t.VectorStoreID, t.RAGID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down Bedrock RAG pipeline: %w", err))
			slog.Error("Failed to tear down Bedrock RAG pipeline", "error", err)
		} else {
			slog.Info("Bedrock RAG pipeline torn down successfully")
		}
	}

	// Report any teardown errors
	if len(teardownErrors) > 0 {
		slog.Error("Bedrock teardown completed with errors", "errorCount", len(teardownErrors))
		// Return the first error, but log all errors
		for i, err := range teardownErrors {
			slog.Error("Bedrock teardown error", "index", i+1, "error", err)
		}
		return teardownErrors[0]
	}

	slog.Info("Bedrock teardown workflow completed successfully")
	return nil
}
