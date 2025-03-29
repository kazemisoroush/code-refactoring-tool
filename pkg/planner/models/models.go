// Package models provides the data models for the code analysis tool.
package models

import (
	"sort"
	"strings"
)

// Plan represents a plan to fix the issues in the source code.
type Plan struct {
	Actions []PlannedAction `json:"actions"`
}

// PlannedAction represents a planned action to fix a specific issue in the source code.
type PlannedAction struct {
	FilePath string       `json:"file_path"`
	Edits    []EditRegion `json:"edits"`
	Reason   string       `json:"reason"`
}

// EditRegion represents a region in the source code that needs to be edited.
type EditRegion struct {
	StartLine   int      `json:"start_line"`
	EndLine     int      `json:"end_line"`
	Replacement []string `json:"replacement"`
}

// Normalize merges and sorts planned actions per file and within each file.
func (plan Plan) Normalize() Plan {
	merged := make(map[string]*PlannedAction)

	for _, action := range plan.Actions {
		if existing, ok := merged[action.FilePath]; ok {
			// Merge edits
			existing.Edits = append(existing.Edits, action.Edits...)

			// Merge reasons if different
			if !strings.Contains(existing.Reason, action.Reason) {
				existing.Reason = strings.TrimSpace(existing.Reason + "\n" + action.Reason)
			}
		} else {
			// Clone action
			cp := action
			merged[action.FilePath] = &cp
		}
	}

	var result []PlannedAction

	for _, action := range merged {
		// Sort edits within each file from bottom to top
		sort.SliceStable(action.Edits, func(i, j int) bool {
			return action.Edits[i].StartLine > action.Edits[j].StartLine
		})
		result = append(result, *action)
	}

	// Optional: sort actions by file path to keep output deterministic
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].FilePath < result[j].FilePath
	})

	return Plan{Actions: result}
}
