// Package models provides minimal relationship types for task execution
package models

// TaskWithFullContext represents a task with related resources for execution
// This is a minimal version - only includes what's actually needed for task execution
type TaskWithFullContext struct {
	Task
	// Related resources loaded when needed for execution
	Project  *Project  `json:"project,omitempty" db:"-"`
	Agent    *Agent    `json:"agent,omitempty" db:"-"`
	Codebase *Codebase `json:"codebase,omitempty" db:"-"`
}

// CodebaseWithContext represents a codebase with project context for RAG creation
// This is a minimal version - only includes what the factory needs
type CodebaseWithContext struct {
	Codebase
	Project *Project `json:"project,omitempty" db:"-"`
}
