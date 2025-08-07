// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
)

// TeardownLocalSetupWorkflow represents a workflow for tearing down local AI resources
type TeardownLocalSetupWorkflow struct {
	repository   codebase.Codebase
	ragBuilder   builder.RAGBuilder
	agentBuilder builder.AgentBuilder

	// Resource IDs to tear down
	vectorStoreID string
	ragID         string
	agentID       string
	agentVersion  string
}

// NewTeardownLocalSetupWorkflow creates a new TeardownLocalSetupWorkflow instance
func NewTeardownLocalSetupWorkflow(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &TeardownLocalSetupWorkflow{
		repository:   repo,
		ragBuilder:   ragBuilder,
		agentBuilder: agentBuilder,
	}, nil
}

// NewTeardownLocalSetupWorkflowWithResources creates a new TeardownLocalSetupWorkflow instance with resource IDs
func NewTeardownLocalSetupWorkflowWithResources(
	repo codebase.Codebase,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
	vectorStoreID, ragID, agentID, agentVersion string,
) (Workflow, error) {
	return &TeardownLocalSetupWorkflow{
		repository:    repo,
		ragBuilder:    ragBuilder,
		agentBuilder:  agentBuilder,
		vectorStoreID: vectorStoreID,
		ragID:         ragID,
		agentID:       agentID,
		agentVersion:  agentVersion,
	}, nil
}

// SetResourceIDs sets the resource IDs for local teardown
func (t *TeardownLocalSetupWorkflow) SetResourceIDs(vectorStoreID, ragID, agentID, agentVersion string) {
	t.vectorStoreID = vectorStoreID
	t.ragID = ragID
	t.agentID = agentID
	t.agentVersion = agentVersion
}

// Run implements Workflow for tearing down local resources.
func (t *TeardownLocalSetupWorkflow) Run(ctx context.Context) error {
	slog.Info("Running local teardown workflow")

	defer func() {
		err := t.repository.Cleanup()
		if err != nil {
			slog.Error("failed to cleanup repository", "error", err)
		}
	}()

	// Track teardown errors to report at the end
	var teardownErrors []error

	// 1. Tear down the local agent if we have the IDs
	if t.agentID != "" && t.ragID != "" {
		slog.Info("Tearing down local agent", "agentID", t.agentID)
		if err := t.agentBuilder.TearDown(ctx, t.agentID, t.agentVersion, t.ragID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down local agent: %w", err))
			slog.Error("Failed to tear down local agent", "error", err)
		} else {
			slog.Info("Local agent torn down successfully")
		}
	}

	// 2. Tear down the local RAG pipeline if we have the IDs
	if t.vectorStoreID != "" && t.ragID != "" {
		slog.Info("Tearing down local RAG pipeline", "ragID", t.ragID)
		if err := t.ragBuilder.TearDown(ctx, t.vectorStoreID, t.ragID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down local RAG pipeline: %w", err))
			slog.Error("Failed to tear down local RAG pipeline", "error", err)
		} else {
			slog.Info("Local RAG pipeline torn down successfully")
		}
	}

	// Report any teardown errors
	if len(teardownErrors) > 0 {
		slog.Error("Local teardown completed with errors", "errorCount", len(teardownErrors))
		// Return the first error, but log all errors
		for i, err := range teardownErrors {
			slog.Error("Local teardown error", "index", i+1, "error", err)
		}
		return teardownErrors[0]
	}

	slog.Info("Local teardown workflow completed successfully")
	return nil
}
