// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/kazemisoroush/code-refactor-tool/pkg/patcher"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactor-tool/pkg/repository"
)

// Workflow represents a code analysis workflow
type Workflow struct {
	Config     config.Config
	Analyzer   analyzer.Analyzer
	Repository repository.Repository
	Planner    planner.Planner
	Patcher    patcher.Patcher
}

// NewWorkflow creates a new Workflow instance
func NewWorkflow(
	cfg config.Config,
	analyzer analyzer.Analyzer,
	repo repository.Repository,
	planner planner.Planner,
	patcher patcher.Patcher,
) (*Workflow, error) {
	return &Workflow{
		Config:     cfg,
		Analyzer:   analyzer,
		Repository: repo,
		Planner:    planner,
		Patcher:    patcher,
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

	if len(issues) == 0 {
		log.Println("No issues found.")
		return nil
	}

	plan, err := o.Planner.Plan(ctx, path, issues)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	err = o.Patcher.Patch(path, plan)
	if err != nil {
		return fmt.Errorf("failed to patch code: %w", err)
	}

	err = o.Repository.CheckoutBranch("fix/refactor")
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	err = o.Repository.Commit("Refactor code")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	err = o.Repository.Push()
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}
