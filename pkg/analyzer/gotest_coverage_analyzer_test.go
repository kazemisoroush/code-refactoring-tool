package analyzer_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoTestCoverageAnalyzer(t *testing.T) {
	// Arrange
	dir := t.TempDir()

	// Initialize go module
	cmd := exec.Command("go", "mod", "init", "example.com/testmod")
	cmd.Dir = dir
	err := cmd.Run()
	require.NoError(t, err, "go mod init should succeed")

	// Create main.go
	mainPath := filepath.Join(dir, "main.go")
	err = os.WriteFile(mainPath, []byte(`package main

func main() {
	doSomething()
}

func doSomething() {}
`), 0644)
	require.NoError(t, err)

	// Create a test file
	testPath := filepath.Join(dir, "main_test.go")
	err = os.WriteFile(testPath, []byte(`package main

import "testing"

func TestDoSomething(t *testing.T) {
	doSomething()
}
`), 0644)
	require.NoError(t, err)

	// Act
	a := analyzer.NewGoTestCoverageAnalyzer()
	result, err := a.AnalyzeCode(dir)
	require.NoError(t, err, "AnalyzeCode should not return an error")

	issues, extractErr := a.ExtractIssues(result)

	// Assert
	assert.NoError(t, extractErr)
	assert.IsType(t, []models.CodeIssue{}, issues)
}
