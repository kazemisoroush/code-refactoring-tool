package planner

import (
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
)

// AIPlanner is the interface that wraps the PlanFixes method.
type AIPlanner struct {
}

// NewAIPlanner constructor.
func NewAIPlanner() Planner {
	return &AIPlanner{}
}

// PlanFixes implements Planner.
func (a *AIPlanner) PlanFixes(_ string) (models.FixPlan, error) {
	panic("unimplemented")
}
