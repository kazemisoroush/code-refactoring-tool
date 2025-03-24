package analyzer

import (
	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

// CodeAnalyzer is an interface for analyzing Go code
//
//go:generate mockgen -destination=./mocks/mock_analyzer.go -mock_names=CodeAnalyzer=MockCodeAnalyzer -package=mocks . CodeAnalyzer
type CodeAnalyzer interface {
	// AnalyzeCode runs the code analysis tool on the provided source path
	AnalyzeCode(sourcePath string) (models.AnalysisResult, error)

	// ExtractMetrics extracts code metrics from the analysis result
	ExtractMetrics(result models.AnalysisResult) (models.CodeMetrics, error)

	// GenerateReport generates a report from the code metrics
	GenerateReport(metrics models.CodeMetrics) models.Report
}
