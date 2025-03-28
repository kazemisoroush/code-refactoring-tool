package patcher

import (
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner/models"
)

// Patcher represents a code patcher
//
//go:generate mockgen -destination=./mocks/mock_patcher.go -mock_names=Patcher=MockPatcher -package=mocks . Patcher
type Patcher interface {
	// Patch creates a patch to fix the issues in the source code
	Patch(projectSourcePath string, plan models.Plan) error
}
