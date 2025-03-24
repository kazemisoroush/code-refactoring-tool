package analyzer

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

type GoAnalyzer struct{}

// NewGoAnalyzer creates a new GoAnalyzer.
func NewGoAnalyzer() (CodeAnalyzer, error) {
	// Check the golangci-lint cli exists...
	output, err := exec.Command("golangci-lint", "--version").Output()
	if err != nil {
		return GoAnalyzer{}, fmt.Errorf("golangci-lint not found: %v", err)
	}

	fmt.Println("golangci-lint version:", string(output))

	return GoAnalyzer{}, nil
}

// AnalyzeCode implements CodeAnalyzer.
func (g GoAnalyzer) AnalyzeCode(sourcePath string) (models.AnalysisResult, error) {
	output, err := exec.Command(
		"golangci-lint",
		"run",
		"--out-format",
		"json",
		sourcePath,
	).Output()
	if output == nil {
		return models.AnalysisResult{}, fmt.Errorf("golangci-lint returned no output")
	}
	if err != nil {
		fmt.Println("Error running golangci-lint:", err)
		fmt.Println("Output:", string(output))
	}

	return models.AnalysisResult{RawOutput: string(output)}, nil
}

// ExtractMetrics implements CodeAnalyzer.
func (g GoAnalyzer) ExtractMetrics(result models.AnalysisResult) (models.CodeMetrics, error) {
	golangCILintReport := &models.GolangCILintReport{}
	err := json.Unmarshal([]byte(result.RawOutput), golangCILintReport)
	if err != nil {
		return models.CodeMetrics{}, fmt.Errorf("error unmarshalling golangci-lint report: %v %s", err, result.RawOutput)
	}

	return models.CodeMetrics{
		CyclomaticComplexity: golangCILintReport.GetCyclomaticComplexity(),
		DuplicateCode:        golangCILintReport.GetDuplicateCode(),
		TestCoverage:         golangCILintReport.GetTestCoverage(),
		FunctionCount:        golangCILintReport.GetFunctionCount(),
		LongFunctions:        golangCILintReport.GetLongFunctions(),
		DeadCodeCount:        golangCILintReport.GetDeadCodeCount(),
	}, nil
}

// GenerateReport implements CodeAnalyzer.
func (g GoAnalyzer) GenerateReport(metrics models.CodeMetrics) models.Report {
	suggestions := []string{}
	if metrics.CyclomaticComplexity > 15 {
		suggestions = append(suggestions, "Reduce cyclomatic complexity by refactoring functions.")
	}
	if metrics.DuplicateCode > 10 {
		suggestions = append(suggestions, "Remove duplicate code blocks.")
	}

	return models.Report{
		Language:    "Go",
		CodeMetrics: metrics,
		Suggestions: suggestions,
	}
}
