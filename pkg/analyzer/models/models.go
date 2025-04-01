package models

// ToolName tool name string.
type ToolName string

const (
	// ToolNameGoBuild tool name for go build tool.
	ToolNameGoBuild = "go build"

	// ToolNameGolangCI tool name for golang-ci tool.
	ToolNameGolangCI = "golangci-lint"

	// ToolNameGoTest tool name for go test tool.
	ToolNameGoTest = "go test"
)

// IssueType issue type string.
type IssueType string

const (
	// IssueTypeLinter Linter Issue Type.
	IssueTypeLinter IssueType = "linter"

	// IssueTypeBuild Build Issue Type.
	IssueTypeBuild IssueType = "build"

	// IssueTypeTest Test Issue Type.
	IssueTypeTest IssueType = "test"

	// IssueTypeCoverage Test Coverage Issue Type.
	IssueTypeCoverage IssueType = "coverage"
)

// CodeIssue represents a standardized structure for linter findings across different languages.
type CodeIssue struct {
	Tool          ToolName  `json:"tool"`                     // Name of the linter tool (e.g., golangci-lint, pylint)
	Type          IssueType `json:"type"`                     // build / test / linter
	RuleID        string    `json:"rule_id"`                  // Identifier for the violated rule
	Message       string    `json:"message"`                  // Description of the issue
	FilePath      string    `json:"file_path,omitempty"`      // Path to the file containing the issue
	Line          int       `json:"line,omitempty"`           // Line number where the issue occurs
	Column        int       `json:"column,omitempty"`         // Column number (optional)
	SourceSnippet []string  `json:"source_snippet,omitempty"` // Code snippet related to the issue
	Suggestions   []string  `json:"suggestions,omitempty"`    // Recommended fixes or improvements
}

// AnalysisResult represents the result of a code analysis.
type AnalysisResult struct {
	RawOutput string
	Errors    []string
}
