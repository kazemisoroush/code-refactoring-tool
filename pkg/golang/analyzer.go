package golang

import "github.com/kazemisoroush/code-refactor-tool/pkg/golang/models"

//go:generate mockgen -destination=mocks/mock_analyzer.go -package=mocks github.com/braddle/go-tools/pkg/golang CodeAnalyzer
type CodeAnalyzer interface {
	AnalyzeCode(sourcePath string) (models.AnalysisResult, error)
	ExtractMetrics(result models.AnalysisResult) (models.CodeMetrics, error)
	GenerateReport(metrics models.CodeMetrics) models.Report
}
