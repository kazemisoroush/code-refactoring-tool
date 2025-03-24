package fixer

import (
	"fmt"
	"os/exec"
)

// Fixer is the interface that wraps the Fix method.
type GolangCIFixer struct {
}

// NewGolangCIFixer creates a new GolangCIFixer.
func NewGolangCIFixer() Fixer {
	return &GolangCIFixer{}
}

// FixIssues implements Fixer.
func (f *GolangCIFixer) FixIssues(sourcePath string) error {
	fmt.Println("🔧 Running static fixes...")
	cmd := exec.Command(
		"golangci-lint",
		"run",
		"--fix",
		"--out-format",
		"json",
		sourcePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fix issues: %w\nOutput: %s", err, output)
	}

	return nil
}
