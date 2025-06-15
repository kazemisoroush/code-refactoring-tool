// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/patcher"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// FixWorkflow represents a code analysis workflow
type FixWorkflow struct {
	Config     config.Config
	Analyzers  []analyzer.Analyzer
	Repository repository.Repository
	Planner    planner.Planner
	Patcher    patcher.Patcher
}

// NewFixWorkflow creates a new Workflow instance
func NewFixWorkflow(
	cfg config.Config,
	repo repository.Repository,
	planner planner.Planner,
	patcher patcher.Patcher,
	analyzers ...analyzer.Analyzer,
) (Workflow, error) {
	return &FixWorkflow{
		Config:     cfg,
		Repository: repo,
		Planner:    planner,
		Patcher:    patcher,
		Analyzers:  analyzers,
	}, nil
}

// Run executes the code analysis workflow
func (o *FixWorkflow) Run(ctx context.Context) error {
	log.Println("Running workflow...")

	// defer func() {
	// 	err := o.Repository.Cleanup()
	// 	if err != nil {
	// 		log.Printf("failed to cleanup repository: %v", err)
	// 	}
	// }()

	// Clone the repository
	err := o.Repository.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	path := o.Repository.GetPath()

	allIssues := []models.CodeIssue{}

	for _, a := range o.Analyzers {
		log.Printf("Running analyzer: %T", a)

		result, err := a.AnalyzeCode(path)
		if err != nil {
			log.Printf("Analyzer %T failed: %v", a, err)
			// still continue to extract issues
		}

		issues, err := a.ExtractIssues(result)
		if err != nil {
			return fmt.Errorf("failed to extract issues from %T: %w", a, err)
		}
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) == 0 {
		log.Println("No issues found.")
		return nil
	}

	plan, err := o.Planner.Plan(ctx, path, []models.CodeIssue{allIssues[0]})
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

	err = o.Repository.Push()
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	output, err := o.Repository.CreatePR(
		plan.Change.Title,
		plan.Change.Description,
		"fix/linter",
		"main",
	)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}
	log.Println("PR created:", output)

	return nil
}
