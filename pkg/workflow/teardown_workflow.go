// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// TeardownWorkflow represents a workflow for tearing down AI resources
type TeardownWorkflow struct {
	Config       config.Config
	Repository   repository.Repository
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs to tear down
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewTeardownWorkflow creates a new TeardownWorkflow instance
func NewTeardownWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &TeardownWorkflow{
		Config:       cfg,
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// NewTeardownWorkflowWithResources creates a new TeardownWorkflow instance with resource IDs
func NewTeardownWorkflowWithResources(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
	vectorStoreID, ragID, agentID, agentVersion string,
) (Workflow, error) {
	return &TeardownWorkflow{
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

// SetResourceIDs sets the resource IDs for teardown
func (t *TeardownWorkflow) SetResourceIDs(vectorStoreID, ragID, agentID, agentVersion string) {
	t.VectorStoreID = vectorStoreID
	t.RAGID = ragID
	t.AgentID = agentID
	t.AgentVersion = agentVersion
}

// Run implements Workflow.
func (t *TeardownWorkflow) Run(ctx context.Context) error {
	log.Println("Running teardown workflow...")

	// Track teardown errors to report at the end
	var teardownErrors []error

	// 1. Tear down the agent if we have the IDs
	if t.AgentID != "" && t.RAGID != "" {
		log.Printf("Tearing down agent: %s", t.AgentID)
		if err := t.AgentBuilder.TearDown(ctx, t.AgentID, t.AgentVersion, t.RAGID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down agent: %w", err))
			log.Printf("❌ Failed to tear down agent: %v", err)
		} else {
			log.Printf("✅ Agent torn down successfully")
		}
	}

	// 2. Tear down the RAG pipeline if we have the IDs
	if t.VectorStoreID != "" && t.RAGID != "" {
		log.Printf("Tearing down RAG pipeline: %s", t.RAGID)
		if err := t.RAGBuilder.TearDown(ctx, t.VectorStoreID, t.RAGID); err != nil {
			teardownErrors = append(teardownErrors, fmt.Errorf("failed to tear down RAG pipeline: %w", err))
			log.Printf("❌ Failed to tear down RAG pipeline: %v", err)
		} else {
			log.Printf("✅ RAG pipeline torn down successfully")
		}
	}

	// 3. Clean up the repository
	log.Println("Cleaning up repository...")
	if err := t.Repository.Cleanup(); err != nil {
		teardownErrors = append(teardownErrors, fmt.Errorf("failed to clean up repository: %w", err))
		log.Printf("❌ Failed to clean up repository: %v", err)
	} else {
		log.Printf("✅ Repository cleaned up successfully")
	}

	// Report any teardown errors
	if len(teardownErrors) > 0 {
		log.Printf("❌ Teardown completed with %d errors", len(teardownErrors))
		// Return the first error, but log all errors
		for i, err := range teardownErrors {
			log.Printf("  Error %d: %v", i+1, err)
		}
		return teardownErrors[0]
	}

	log.Println("✅ Teardown workflow completed successfully")
	return nil
}
