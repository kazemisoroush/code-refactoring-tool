// Package workflow defines workflow interface.
package workflow

import (
	"context"
)

// Workflow interface.
//
//go:generate mockgen -destination=./mocks/mock_workflow.go -mock_names=Workflow=MockWorkflow -package=mocks . Workflow
type Workflow interface {
	// Run workflow.
	Run(ctx context.Context) error
}
