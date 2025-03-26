package planner_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	agent_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/agent/mocks"
	analyzer_models "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIPlanner_NoIssues(t *testing.T) {
	// Arrange
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agnt := agent_mocks.NewMockAgent(ctrl)

	p := planner.NewAIPlanner(agnt)

	issues := []analyzer_models.LinterIssue{}

	// Act
	_, err := p.Plan(ctx, "sourcePath", issues)

	// Assert
	assert.NoError(t, err, "Plan should not return an error")
}

func TestAIPlanner_Issues(t *testing.T) {
	// Arrange
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	agnt := agent_mocks.NewMockAgent(ctrl)
	expectedPlanActions := []models.PlannedAction{
		{
			FilePath: "filePath",
			Edits: []models.EditRegion{
				{
					StartLine:   1,
					EndLine:     1,
					Replacement: []string{"replacement"},
				},
			},
			Reason: "reason",
		},
	}
	expectedPlanActionsBytes, err := json.Marshal(expectedPlanActions)
	require.NoError(t, err, "MarshalJSON should not return an error")

	agnt.EXPECT().Ask(ctx, gomock.Any()).Return(string(expectedPlanActionsBytes), nil)

	p := planner.NewAIPlanner(agnt)

	issues := []analyzer_models.LinterIssue{
		{
			LinterName:    "",
			RuleID:        "",
			Message:       "",
			FilePath:      "",
			Line:          1,
			Column:        1,
			SourceSnippet: []string{"snippet1", "snippet2"},
			Suggestions:   []string{"suggestion1", "suggestion2"},
		},
	}

	// Act
	plan, err := p.Plan(ctx, "sourcePath", issues)

	// Assert
	assert.NoError(t, err, "Plan should not return an error")
	assert.Len(t, plan.Actions, len(expectedPlanActions))
	for i, action := range plan.Actions {
		if !reflect.DeepEqual(action, expectedPlanActions[i]) {
			t.Errorf("expected %v, got %v", expectedPlanActions[i], action)
		}
	}
}
