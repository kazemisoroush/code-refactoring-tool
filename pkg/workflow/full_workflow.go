// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/patcher"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner"
)

// FullWorkflow represents a code analysis workflow
type FullWorkflow struct {
	Config     config.Config
	Analyzers  []analyzer.Analyzer
	Repository codebase.Codebase
	Planner    planner.Planner
	Patcher    patcher.Patcher
}

// NewFullWorkflow creates a new Workflow instance
func NewFullWorkflow(
	cfg config.Config,
	repo codebase.Codebase,
	planner planner.Planner,
	patcher patcher.Patcher,
	analyzers ...analyzer.Analyzer,
) (Workflow, error) {
	return &FullWorkflow{
		Config:     cfg,
		Repository: repo,
		Planner:    planner,
		Patcher:    patcher,
		Analyzers:  analyzers,
	}, nil
}

// Run executes the code analysis workflow
func (o *FullWorkflow) Run(ctx context.Context) error {
	slog.Info("Running workflow")

	defer func() {
		err := o.Repository.Cleanup()
		if err != nil {
			slog.Error("failed to cleanup repository", "error", err)
		}
	}()

	// Clone the repository
	err := o.Repository.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	path := o.Repository.GetPath()

	allIssues := []models.CodeIssue{}

	for _, a := range o.Analyzers {
		slog.Info("Running analyzer", "type", fmt.Sprintf("%T", a))

		result, err := a.AnalyzeCode(path)
		if err != nil {
			slog.Warn("Analyzer failed", "type", fmt.Sprintf("%T", a), "error", err)
			// still continue to extract issues
		}

		issues, err := a.ExtractIssues(result)
		if err != nil {
			return fmt.Errorf("failed to extract issues from %T: %w", a, err)
		}
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) == 0 {
		slog.Info("No issues found")
		return nil
	}

	plan, err := o.Planner.Plan(ctx, path, allIssues)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	err = o.Repository.CheckoutBranch("fix/linter")
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	err = o.Patcher.Patch(path, plan)
	if err != nil {
		return fmt.Errorf("failed to patch code: %w", err)
	}

	err = o.Repository.Commit("Refactor code")
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	err = o.Repository.Push(ctx)
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	output, err := o.Repository.CreatePR(
		ctx,
		plan.Change.Title,
		plan.Change.Description,
		"fix/linter",
		"main",
	)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	slog.Info("PR created", "output", output)

	return nil
}
