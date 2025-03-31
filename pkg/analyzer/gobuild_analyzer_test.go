package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoBuildAnalyzer_ValidCode(t *testing.T) {
	// Arrange
	dir := t.TempDir()

	// create go.mod to make it a Go module
	goModPath := filepath.Join(dir, "go.mod")
	err := os.WriteFile(goModPath, []byte(`module testmod`), 0644)
	require.NoError(t, err)

	mainPath := filepath.Join(dir, "main.go")
	err = os.WriteFile(mainPath, []byte(`package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
}`), 0644)
	require.NoError(t, err)

	// Act
	a := analyzer.NewGoBuildAnalyzer()
	result, err := a.AnalyzeCode(dir)
	issues, extractErr := a.ExtractIssues(result)

	// Assert
	assert.NoError(t, err, "go build should not return an error for valid code")
	assert.NoError(t, extractErr)
	assert.Len(t, issues, 0)
}

func TestGoBuildAnalyzer_InvalidCode(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	mainPath := filepath.Join(dir, "main.go")
	err := os.WriteFile(mainPath, []byte(`package main

func main() {
	fmt.Println("missing import")
}`), 0644)
	require.NoError(t, err)

	// Act
	a := analyzer.NewGoBuildAnalyzer()
	result, err := a.AnalyzeCode(dir)
	issues, extractErr := a.ExtractIssues(result)

	// Assert
	assert.Error(t, err, "go build should fail for invalid code")
	assert.NoError(t, extractErr)
	assert.Len(t, issues, 2)
}
