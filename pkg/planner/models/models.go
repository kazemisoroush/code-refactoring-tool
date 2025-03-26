// Package models provides the data models for the code analysis tool.
package models

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
