package fixer

import (
	"context"

	analyzerModels "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

// Fixer is the interface that wraps the PlanFixes method.
//
//go:generate mockgen -destination=./mocks/mock_fixer.go -mock_names=Fixer=MockFixer -package=mocks . Fixer
type Fixer interface {
	// Fix fixes the code in the provided source path
	Fix(ctx context.Context, sourcePath string, issues []analyzerModels.LinterIssue) error
}
