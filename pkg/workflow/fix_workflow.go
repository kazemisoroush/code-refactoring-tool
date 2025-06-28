// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// FixWorkflow represents a code analysis workflow
type FixWorkflow struct {
	Config     config.Config
	Repository repository.Repository
	RAGBuilder ai.RAGBuilder
}

// NewFixWorkflow creates a new Workflow instance
func NewFixWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder ai.RAGBuilder,
) (Workflow, error) {
	return &FixWorkflow{
		Config:     cfg,
		Repository: repo,
		RAGBuilder: ragBuilder,
	}, nil
}

// Run executes the code analysis workflow
func (o *FixWorkflow) Run(ctx context.Context) error {
	log.Println("Running workflow...")

	ragMetadata, err := o.RAGBuilder.Build(ctx, o.Repository)
	if err != nil {
		return fmt.Errorf("failed to build RAG pipeline: %w", err)
	}
	log.Printf("RAG pipeline built successfully: %+v", ragMetadata)

	return nil
}
