package patcher_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/patcher"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner/models"
	"github.com/stretchr/testify/require"
)

func TestFilePatcher_Patch(t *testing.T) {
	// Arrange
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test file inside temp directory
	originalContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`

	relativePath := "main.go"
	fullPath := filepath.Join(tmpDir, relativePath)

	err := os.WriteFile(fullPath, []byte(originalContent), 0644)
	require.NoError(t, err)

	// Define a patch plan to replace Println line
	plan := models.Plan{
		Actions: []models.PlannedAction{
			{
				FilePath: relativePath,
				Edits: []models.EditRegion{
					{
						StartLine:   5,
						EndLine:     6,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "update greeting",
			},
		},
	}

	patcher := patcher.NewFilePatcher()

	// Act
	err = patcher.Patch(tmpDir, plan)
	require.NoError(t, err)

	// Assert
	// Read back the file and verify the change
	modifiedContentBytes, err := os.ReadFile(fullPath)
	require.NoError(t, err)

	modifiedContent := string(modifiedContentBytes)
	require.Contains(t, modifiedContent, `fmt.Println("Hello, AI World!")`)
	require.NotContains(t, modifiedContent, `fmt.Println("Hello, world!")`)
}
