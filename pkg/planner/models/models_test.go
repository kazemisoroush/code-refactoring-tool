package models_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONStructConversion(t *testing.T) {
	// Arrange
	schemaExample := models.PlannedAction{}

	// Act
	schemaBytes, err := json.MarshalIndent(schemaExample, "", "  ")
	require.NoError(t, err)
	s := string(schemaBytes)

	// Assert
	assert.NotEmpty(t, s)
}

func TestSortPlan(t *testing.T) {
	// Arrange
	plan := models.Plan{
		Actions: []models.PlannedAction{
			{
				FilePath: "a.go",
				Edits: []models.EditRegion{
					{
						StartLine:   3,
						EndLine:     4,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
					{
						StartLine:   10,
						EndLine:     10,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "reason1",
			},
			{
				FilePath: "b.go",
				Edits: []models.EditRegion{
					{
						StartLine:   5,
						EndLine:     10,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "reason2",
			},
			{
				FilePath: "a.go",
				Edits: []models.EditRegion{
					{
						StartLine:   1,
						EndLine:     2,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "reason3",
			},
		},
	}
	expectedPlan := models.Plan{
		Actions: []models.PlannedAction{
			{
				FilePath: "a.go",
				Edits: []models.EditRegion{
					{
						StartLine:   10,
						EndLine:     10,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
					{
						StartLine:   3,
						EndLine:     4,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
					{
						StartLine:   1,
						EndLine:     2,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "reason1\nreason3",
			},
			{
				FilePath: "b.go",
				Edits: []models.EditRegion{
					{
						StartLine:   5,
						EndLine:     10,
						Replacement: []string{`	fmt.Println("Hello, AI World!")`},
					},
				},
				Reason: "reason2",
			},
		},
	}

	// Act
	normalizedPlan := plan.Normalize()

	// Assert
	if !reflect.DeepEqual(normalizedPlan, expectedPlan) {
		t.Errorf("SortPlan() = %v, want %v", normalizedPlan, expectedPlan)
	}
}
