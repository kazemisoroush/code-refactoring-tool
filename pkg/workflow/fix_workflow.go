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

// FixWorkflow represents a code analysis workflow
type FixWorkflow struct {
	Config       config.Config
	Repository   repository.Repository
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder
}

// NewFixWorkflow creates a new Workflow instance
func NewFixWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &FixWorkflow{
		Config:       cfg,
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// Run executes the code analysis workflow
func (o *FixWorkflow) Run(ctx context.Context) error {
	log.Println("Running workflow...")

	// Build the RAG
	ragID, err := o.RAGBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build RAG pipeline: %w", err)
	}
	log.Printf("RAG pipeline built successfully RagID: %+v", ragID)

	// Build the agent
	agentID, agentAliasID, err := o.AgentBuilder.Build(ctx, ragID)
	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}
	log.Printf("Agent built successfully AgentID: %s, AgentAliasID: %s", agentID, agentAliasID)

	// TODO: Ask agent to fix the codebase

	return nil
}
