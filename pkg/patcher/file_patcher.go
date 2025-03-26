package patcher

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
)

type FilePatcher struct {
}

func NewFilePatcher() Patcher {
	return &FilePatcher{}
}

func (p *FilePatcher) Patch(projectSourcePath string, plan models.Plan) error {
	for _, action := range plan.Actions {
		fullPath := filepath.Join(projectSourcePath, action.FilePath)

		// Read original file
		file, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", fullPath, err)
		}

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read file %s: %w", fullPath, err)
		}

		// Sort edits in reverse order to prevent line shifting
		sort.SliceStable(action.Edits, func(i, j int) bool {
			return action.Edits[i].StartLine > action.Edits[j].StartLine
		})

		// Apply edits
		for _, edit := range action.Edits {
			start := edit.StartLine - 1
			end := edit.EndLine

			if start < 0 || end > len(lines) || start >= end {
				return fmt.Errorf("invalid edit range in %s: start=%d end=%d", action.FilePath, edit.StartLine, edit.EndLine)
			}

			// Replace the lines
			lines = append(lines[:start], append(edit.Replacement, lines[end:]...)...)
		}

		// Write updated file
		output := strings.Join(lines, "\n")
		if err := os.WriteFile(fullPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write updated file %s: %w", fullPath, err)
		}
	}

	return nil
}
