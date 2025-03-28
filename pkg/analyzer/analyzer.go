package analyzer

import (
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
)

// Analyzer is an interface for analyzing Go code
//
//go:generate mockgen -destination=./mocks/mock_analyzer.go -mock_names=Analyzer=MockAnalyzer -package=mocks . Analyzer
type Analyzer interface {
	// AnalyzeCode runs the code analysis tool on the provided source path
	AnalyzeCode(sourcePath string) (models.AnalysisResult, error)

	// ExtractIssues extracts code metrics from the analysis result
	ExtractIssues(result models.AnalysisResult) ([]models.LinterIssue, error)
}
