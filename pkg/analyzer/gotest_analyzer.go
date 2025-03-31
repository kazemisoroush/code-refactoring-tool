// Package analyzer implements test analyzer.
package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
)

// GoTestAnalyzer runs `go test` to validate tests.
type GoTestAnalyzer struct{}

// NewGoTestAnalyzer creates a new test analyzer.
func NewGoTestAnalyzer() Analyzer {
	return GoTestAnalyzer{}
}

// AnalyzeCode runs `go test ./...` and collects output.
func (g GoTestAnalyzer) AnalyzeCode(sourcePath string) (models.AnalysisResult, error) {
	cmd := exec.Command("go", "test", "./...", "-v", "-json")
	cmd.Dir = sourcePath
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	result := models.AnalysisResult{
		RawOutput: stdout.String(),
	}
	if err != nil {
		// command failed, maybe test failed
		return result, fmt.Errorf("go test failed: %w", err)
	}

	return result, nil
}

// ExtractIssues parses test output and generates issues if any.
func (g GoTestAnalyzer) ExtractIssues(result models.AnalysisResult) ([]models.CodeIssue, error) {
	var issues []models.CodeIssue

	lines := strings.Split(result.RawOutput, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		testIssue := models.GoTestIssue{}
		err := json.Unmarshal([]byte(line), &testIssue)
		if err != nil {
			continue
		}

		if testIssue.Action == "fail" {
			issues = append(issues, models.CodeIssue{
				Tool:    models.ToolNameGoTest,
				Type:    models.IssueTypeTest,
				RuleID:  testIssue.Action,
				Message: line,
			})
		}
	}

	return issues, nil
}
