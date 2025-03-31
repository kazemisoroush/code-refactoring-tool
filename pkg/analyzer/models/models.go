package models

// AnalysisResult represents the result of a code analysis.
type AnalysisResult struct {
	RawOutput string
	Errors    []string
}

// CodeIssue represents a standardized structure for linter findings across different languages.
type CodeIssue struct {
	Tool          string   `json:"tool"`             // Name of the linter tool (e.g., golangci-lint, pylint)
	RuleID        string   `json:"rule_id"`          // Identifier for the violated rule
	Message       string   `json:"message"`          // Description of the issue
	FilePath      string   `json:"file_path"`        // Path to the file containing the issue
	Line          int      `json:"line"`             // Line number where the issue occurs
	Column        int      `json:"column,omitempty"` // Column number (optional)
	SourceSnippet []string `json:"source_snippet"`   // Code snippet related to the issue
	Suggestions   []string `json:"suggestions"`      // Recommended fixes or improvements
}

// GoBuildIssues Issues from go build CLI.
type GoBuildIssues struct {
	ImportPath string `json:"ImportPath"`
	Action     string `json:"Action"`
	Output     string `json:"Output"`
}
