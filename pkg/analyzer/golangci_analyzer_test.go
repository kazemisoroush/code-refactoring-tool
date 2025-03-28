package analyzer_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractIssues(t *testing.T) {
	// Arrange
	golangCILintReport := &models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{
				FromLinter:  "typecheck",
				Text:        ": # github.com/kazemisoroush/code-refactoring-tool/pkg/golang\npkg/golang/goanalyzer.go:31:16: undefined: parseCyclo\npkg/golang/goanalyzer.go:32:16: undefined: parseDuplication\npkg/golang/goanalyzer.go:37:25: undefined: calculateCoverage\npkg/golang/goanalyzer.go:38:25: undefined: countFunctions\npkg/golang/goanalyzer.go:39:25: undefined: detectLongFunctions\npkg/golang/goanalyzer.go:40:25: undefined: detectDeadCode",
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
	tmpFile, err := os.CreateTemp("", "dead_code_*.go")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) //nolint:errcheck

	code := `
		package main

		import "fmt"

		func main() {
			fmt.Println(ValidateUser("John", "", 25))
			fmt.Println(ValidateUser("", "Doe", 17))
		}

		func ValidateUser(firstName, lastName string, age int) string {
			if firstName == "" {
				return "Error: First name is required"
			}
			if lastName == "" {
				return "Error: Last name is required"
			}
			if age < 18 {
				return "Error: User must be 18 or older"
			}
			return "Valid user"
		}
	`

	_, err = tmpFile.WriteString(code)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	analyzer, err := analyzer.NewGolangCIAnalyzer()
	require.NoError(t, err, "NewGoAnalyzer should not return an error")

	// Act
	fmt.Println("🔍 Running code analysis...")
	analysisResult, err := analyzer.AnalyzeCode("/tmp")
	require.NoError(t, err, "AnalyzeCode should not return an error")

	codeIssues, err := analyzer.ExtractIssues(analysisResult)

	// Assert
	require.NoError(t, err, "ExtractIssues should not return an error")
	assert.IsType(t, []models.LinterIssue{}, codeIssues)
}
