package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoTestAnalyzer_WithFailingTest(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	mainTest := `
		package main

		import "testing"

		func TestFail(t *testing.T) {
			t.Fatal("this test fails")
		}
	`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main_test.go"), []byte(mainTest), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`module tempmod`), 0644))

	analyzer := analyzer.NewGoTestAnalyzer()

	// Act
	result, err := analyzer.AnalyzeCode(dir)
	assert.Error(t, err)

	issues, extractErr := analyzer.ExtractIssues(result)

	// Assert
	assert.NoError(t, extractErr)
	assert.NotEmpty(t, issues)
	assert.Equal(t, models.IssueTypeTest, issues[0].Type)
}
