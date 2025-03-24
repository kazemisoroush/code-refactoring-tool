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

func TestExtractIssues(t *testing.T) {
	// Arrange
	golangCILintReport := &models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{
				FromLinter:  "typecheck",
				Text:        ": # github.com/kazemisoroush/code-refactor-tool/pkg/golang\npkg/golang/goanalyzer.go:31:16: undefined: parseCyclo\npkg/golang/goanalyzer.go:32:16: undefined: parseDuplication\npkg/golang/goanalyzer.go:37:25: undefined: calculateCoverage\npkg/golang/goanalyzer.go:38:25: undefined: countFunctions\npkg/golang/goanalyzer.go:39:25: undefined: detectLongFunctions\npkg/golang/goanalyzer.go:40:25: undefined: detectDeadCode",
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
	issues, err := goAnalyzer.ExtractIssues(analysisResult)
	require.NoError(t, err, "ExtractIssues should not return an error")

	// Assert
	assert.Len(t, issues, 1)
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

	codeIssues, err := analyzer.ExtractIssues(analysisResult)
	require.NoError(t, err, "ExtractIssues should not return an error")

	// Assert
	assert.Len(t, codeIssues, 1, "There should be 1 code issue")
}
