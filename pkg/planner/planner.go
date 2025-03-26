package planner

import (
	"context"

	analyzerModels "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
)

// Planner
//
//go:generate mockgen -destination=./mocks/mock_planner.go -mock_names=Planner=MockPlanner -package=mocks . Planner
type Planner interface {
	// Plan creates a plan to fix the issues in the source code
	Plan(ctx context.Context, sourcePath string, issues []analyzerModels.LinterIssue) (models.Plan, error)

	// CreatePrompt creates a prompt for the given issue
	CreatePrompt(issue analyzerModels.LinterIssue) (string, error)
}
