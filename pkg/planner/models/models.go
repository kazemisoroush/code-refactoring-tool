package models

// FixPlan represents a plan to fix a code issue
type FixPlan struct {
	File          string
	StartLine     int
	EndLine       int
	SuggestedCode string
	Metadata      map[string]string // e.g. origin: "LLM" or "linter", confidence, ruleName...
}
