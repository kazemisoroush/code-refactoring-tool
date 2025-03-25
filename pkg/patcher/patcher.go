package patcher

import "github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"

// Patcher is the interface that wraps the Patch method.
//
//go:generate mockgen -destination=./mocks/mock_patcher.go -mock_names=Patcher=MockPatcher -package=mocks . Patcher
type Patcher interface {
	// Patch applies the fixes to the provided source path
	Patch(plan []models.FixPlan) error
}
