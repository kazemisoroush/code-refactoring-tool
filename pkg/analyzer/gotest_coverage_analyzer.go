package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
)

// GoTestCoverageAnalyzer class type.
type GoTestCoverageAnalyzer struct{}

// NewGoTestCoverageAnalyzer class constructor.
func NewGoTestCoverageAnalyzer() Analyzer {
	return GoTestCoverageAnalyzer{}
}

// AnalyzeCode Analyze code coverage for golang.
func (g GoTestCoverageAnalyzer) AnalyzeCode(sourcePath string) (models.AnalysisResult, error) {
	coverFile := filepath.Join(sourcePath, "coverage.out")
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	cmd.Dir = sourcePath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return models.AnalysisResult{RawOutput: string(output)}, fmt.Errorf("go test -coverprofile failed: %v", err)
	}

	data, err := os.ReadFile(coverFile)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("failed to read coverprofile: %v", err)
	}

	return models.AnalysisResult{RawOutput: string(data)}, nil
}

// ExtractIssues Extract test coverage issues.
func (g GoTestCoverageAnalyzer) ExtractIssues(result models.AnalysisResult) ([]models.CodeIssue, error) {
	var issues []models.CodeIssue
	scanner := bufio.NewScanner(strings.NewReader(result.RawOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		location := parts[0]
		numStmts, _ := strconv.Atoi(parts[1])
		count, _ := strconv.Atoi(parts[2])

		if count > 0 {
			continue // covered
		}

		fileAndRange := strings.Split(location, ":")
		if len(fileAndRange) != 2 {
			continue
		}
		filePath := fileAndRange[0]
		rangePart := fileAndRange[1]

		startEnd := strings.Split(rangePart, ",")
		if len(startEnd) != 2 {
			continue
		}
		startLine, _ := strconv.Atoi(strings.Split(startEnd[0], ".")[0])
		endLine, _ := strconv.Atoi(strings.Split(startEnd[1], ".")[0])

		issues = append(issues, models.CodeIssue{
			Tool:     models.ToolNameGoTest,
			Type:     models.IssueTypeCoverage,
			FilePath: filePath,
			Line:     startLine,
			Message:  fmt.Sprintf("Uncovered block from line %d to %d (%d statements)", startLine, endLine, numStmts),
		})
	}
	return issues, nil
}
