package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactor-tool/pkg/repository"
)

// Workflow represents a code analysis workflow
type Workflow struct {
	Config     config.Config
	Analyzer   analyzer.Analyzer
	Repository repository.Repository
	Planner    planner.Planner
}

// NewWorkflow creates a new Workflow instance
func NewWorkflow(
	cfg config.Config,
	analyzer analyzer.Analyzer,
	repo repository.Repository,
	planner planner.Planner,
) (*Workflow, error) {
	return &Workflow{
		Config:     cfg,
		Analyzer:   analyzer,
		Repository: repo,
		Planner:    planner,
	}, nil
}

// Run executes the code analysis workflow
func (o *Workflow) Run(ctx context.Context) error {
	log.Println("Running workflow...")

	// Clone the repository
	err := o.Repository.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	path := o.Repository.GetPath()

	// Analyze the code
	analysisResult, err := o.Analyzer.AnalyzeCode(path)
	if err != nil {
		return fmt.Errorf("failed to analyze code: %w", err)
	}

	// Extract issues
	issues, err := o.Analyzer.ExtractIssues(analysisResult)
	if err != nil {
		return fmt.Errorf("failed to extract issues: %w", err)
	}

	_, err = o.Planner.Plan(ctx, path, issues)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return nil
}
