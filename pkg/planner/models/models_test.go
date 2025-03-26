package models_test

import (
	"encoding/json"
	"testing"

	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
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
