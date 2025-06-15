// Package analyzer provides the code analysis functionality.
package analyzer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
)

// GolangCIAnalyzer is a code analyzer that uses golangci-lint.
type GolangCIAnalyzer struct{}

// NewGolangCIAnalyzer creates a new GoAnalyzer.
func NewGolangCIAnalyzer() (Analyzer, error) {
	// Check the golangci-lint cli exists...
	output, err := exec.Command("golangci-lint", "--version").Output()
	if err != nil {
		return GolangCIAnalyzer{}, fmt.Errorf("golangci-lint not found: %v", err)
	}

	slog.Info("golangci-lint version", "output", string(output))

	return GolangCIAnalyzer{}, nil
}

// AnalyzeCode implements CodeAnalyzer.
func (g GolangCIAnalyzer) AnalyzeCode(sourcePath string) (models.AnalysisResult, error) {
	cmd := exec.Command("golangci-lint", "run", "--output.json.path", "stdout")
	cmd.Dir = sourcePath
	output, err := cmd.Output()
	if output == nil {
		return models.AnalysisResult{}, fmt.Errorf("golangci-lint returned no output")
	}
	if err != nil {
		slog.Error("Error running golangci-lint:", "error", err, "output", string(output))
	}

	// Keep only the first line (JSON) and discard any trailing "0 issues." etc.
	outputLines := strings.SplitN(string(output), "\n", 2)
	cleanOutput := outputLines[0]

	return models.AnalysisResult{RawOutput: cleanOutput}, nil
}

// ExtractIssues transforms golangci-lint issues into a universal linter issue format.
func (g GolangCIAnalyzer) ExtractIssues(result models.AnalysisResult) ([]models.CodeIssue, error) {
	var golangCILintReport models.GolangCILintReport
	err := json.Unmarshal([]byte(result.RawOutput), &golangCILintReport)
	if err != nil {
		slog.Info("Raw output", "error", err, "output", result.RawOutput)
		return nil, fmt.Errorf("error unmarshalling golangci-lint report: %v", err)
	}

	var linterIssues []models.CodeIssue
	for _, issue := range golangCILintReport.Issues {
		linterIssue := models.CodeIssue{
			Tool:          models.ToolNameGolangCI,
			Type:          models.IssueTypeLinter,
			RuleID:        issue.FromLinter,
			Message:       issue.Text,
			FilePath:      issue.Pos.Filename,
			Line:          issue.Pos.Line,
			Column:        issue.Pos.Column,
			SourceSnippet: issue.SourceLines,
			Suggestions:   []string{issue.Replacement.Text},
		}
		linterIssues = append(linterIssues, linterIssue)
	}

	return linterIssues, nil
}
