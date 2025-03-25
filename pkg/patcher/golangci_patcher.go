package patcher

import (
	"fmt"
	"os/exec"

	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
)

// GolangCIPatcher is the interface that wraps the PlanFixes method.
type GolangCIPatcher struct {
	sourcePath string
}

// NewGolangCIPatcher creates a new GolangCIPatcher.
func NewGolangCIPatcher(sourcePath string) Patcher {
	return &GolangCIPatcher{
		sourcePath: sourcePath,
	}
}

// PlanFixes fixes the code in the provided source path.
func (f *GolangCIPatcher) Patch(_ []models.FixPlan) error {
	fmt.Println("🔧 Running static fixes...")
	cmd := exec.Command(
		"golangci-lint",
		"run",
		"--fix",
		"--out-format",
		"json",
		f.sourcePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fix issues: %w\nOutput: %s", err, output)
	}

	return nil
}
