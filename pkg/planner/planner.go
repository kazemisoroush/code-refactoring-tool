package planner

import (
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
)

// Planner is the interface that wraps the PlanFixes method.
//
//go:generate mockgen -destination=./mocks/mock_planner.go -mock_names=Planner=MockPlanner -package=mocks . Planner
type Planner interface {
	// PlanFixes fixes the code in the provided source path
	PlanFixes(sourcePath string) (models.FixPlan, error)
}
