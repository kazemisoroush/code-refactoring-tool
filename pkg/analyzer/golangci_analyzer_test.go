package analyzer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractMetrics(t *testing.T) {
	// Arrange
	golangCILintReport := &models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{
				FromLinter:  "typecheck",
				Text:        ": # github.com/kazemisoroush/code-refactor-tool/pkg/golang\npkg/golang/goanalyzer.go:31:16: undefined: parseCyclo\npkg/golang/goanalyzer.go:32:16: undefined: parseDuplication\npkg/golang/goanalyzer.go:37:25: undefined: calculateCoverage\npkg/golang/goanalyzer.go:38:25: undefined: countFunctions\npkg/golang/goanalyzer.go:39:25: undefined: detectLongFunctions\npkg/golang/goanalyzer.go:40:25: undefined: detectDeadCode",
				Severity:    "",
				SourceLines: []string{"package golang"},
				Pos: models.GolangCIPosition{
					Filename: "pkg/golang/analyzer.go",
					Offset:   0,
					Line:     1,
					Column:   0,
				},
				ExpectNoLint:         false,
				ExpectedNoLintLinter: "",
			},
		},
		Report: models.GolangCIReport{
			Linters: []models.GolangCILinter{
				{
					Name: "asasalint",
				},
			},
		},
	}

	jsonMarshalBytes, err := json.Marshal(golangCILintReport)
	require.NoError(t, err, "json.Marshal should not return an error")

	goAnalyzer, err := analyzer.NewGolangCIAnalyzer()
	require.NoError(t, err, "NewGoAnalyzer should not return an error")

	analysisResult := models.AnalysisResult{
		RawOutput: string(jsonMarshalBytes),
		Errors:    []string{},
	}

	// Act
	metrics, err := goAnalyzer.ExtractMetrics(analysisResult)
	require.NoError(t, err, "ExtractMetrics should not return an error")

	// Assert
	assert.Equal(t, 0, metrics.CyclomaticComplexity)
	assert.Equal(t, 0, metrics.DuplicateCode)
	assert.Equal(t, float64(0), metrics.TestCoverage)
	assert.Equal(t, 0, metrics.FunctionCount)
	assert.Equal(t, 0, metrics.LongFunctions)
	assert.Equal(t, 0, metrics.DeadCodeCount)
}

func TestGolangCodeAnalyzer_Integration(t *testing.T) {
	// Arrange
	sourcePath := "./fixtures/dead_code.go"

	analyzer, err := analyzer.NewGolangCIAnalyzer()
	require.NoError(t, err, "NewGoAnalyzer should not return an error")

	// Act
	fmt.Println("🔍 Running code analysis...")
	analysisResult, err := analyzer.AnalyzeCode(sourcePath)
	require.NoError(t, err, "AnalyzeCode should not return an error")

	codeMetrics, err := analyzer.ExtractMetrics(analysisResult)
	require.NoError(t, err, "ExtractMetrics should not return an error")

	report := analyzer.GenerateReport(codeMetrics)

	// Assert
	assert.Equal(t, "Go", report.Language)
	assert.Equal(t, 0, report.CodeMetrics.CyclomaticComplexity)
	assert.Equal(t, 0, report.CodeMetrics.DuplicateCode)
	assert.Equal(t, float64(0), report.CodeMetrics.TestCoverage)
	assert.Equal(t, 0, report.CodeMetrics.FunctionCount)
	assert.Equal(t, 0, report.CodeMetrics.LongFunctions)
	assert.Equal(t, 1, report.CodeMetrics.DeadCodeCount)
}
