package models

import (
	"fmt"
)

// AnalysisResult represents the result of a code analysis.
type AnalysisResult struct {
	RawOutput string
	Errors    []string
}

// CodeMetrics represents the metrics of a code analysis.
type CodeMetrics struct {
	CyclomaticComplexity int
	DuplicateCode        int
	TestCoverage         float64
	FunctionCount        int
	LongFunctions        int
	DeadCodeCount        int
}

// Report represents a code analysis report.
type Report struct {
	Language    string
	CodeMetrics CodeMetrics
	Suggestions []string
}

// ToString returns a string representation of the CodeMetrics struct.
func (r Report) ToString() string {
	return fmt.Sprintf(
		"Cyclomatic Complexity: %d\nDuplicate Code: %d\nTest Coverage: %.2f%%\nFunction Count: %d\nLong Functions: %d\nDead Code Count: %d",
		r.CodeMetrics.CyclomaticComplexity,
		r.CodeMetrics.DuplicateCode,
		r.CodeMetrics.TestCoverage,
		r.CodeMetrics.FunctionCount,
		r.CodeMetrics.LongFunctions,
		r.CodeMetrics.DeadCodeCount,
	)
}
