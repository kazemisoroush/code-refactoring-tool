package models

// AnalysisResult represents the result of a code analysis.
type AnalysisResult struct {
	RawOutput string
	Errors    []string
}

// LinterIssue represents a standardized structure for linter findings across different languages.
type LinterIssue struct {
	LinterName    string   `json:"linter_name"`      // Name of the linter tool (e.g., golangci-lint, pylint)
	RuleID        string   `json:"rule_id"`          // Identifier for the violated rule
	Message       string   `json:"message"`          // Description of the issue
	FilePath      string   `json:"file_path"`        // Path to the file containing the issue
	Line          int      `json:"line"`             // Line number where the issue occurs
	Column        int      `json:"column,omitempty"` // Column number (optional)
	SourceSnippet []string `json:"source_snippet"`   // Code snippet related to the issue
	Suggestions   []string `json:"suggestions"`      // Recommended fixes or improvements
}
