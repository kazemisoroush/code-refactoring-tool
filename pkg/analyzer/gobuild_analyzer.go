// Package analyzer implements go build analyzer.
package analyzer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
)

// GoBuildAnalyzer use go build to analyze the code.
type GoBuildAnalyzer struct{}

// NewGoBuildAnalyzer create new instance.
func NewGoBuildAnalyzer() Analyzer {
	return GoBuildAnalyzer{}
}

// AnalyzeCode using go build CLI.
func (g GoBuildAnalyzer) AnalyzeCode(sourcePath string) (models.AnalysisResult, error) {
	cmd := exec.Command("go", "build", "-json", "./...")
	cmd.Dir = sourcePath
	output, err := cmd.Output()
	result := models.AnalysisResult{RawOutput: string(output)}
	if err != nil {
		return result, fmt.Errorf("error running go build: %v", err)
	}

	return result, nil
}

// ExtractIssues for the go build analyzer.
func (g GoBuildAnalyzer) ExtractIssues(result models.AnalysisResult) ([]models.CodeIssue, error) {
	if result.RawOutput == "" {
		return nil, nil
	}
	rows := strings.Split(result.RawOutput, "\n")
	issues := []models.CodeIssue{}

	for _, row := range rows {
		if row == "" {
			continue
		}
		buildIssue := models.GoBuildIssues{}
		err := json.Unmarshal([]byte(row), &buildIssue)
		if err != nil {
			continue
		}
		issue := models.CodeIssue{
			Tool:     "go build",
			RuleID:   buildIssue.Action,
			Message:  buildIssue.Output,
			FilePath: buildIssue.ImportPath,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}
