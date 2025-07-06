package planner_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	agent_mocks "github.com/kazemisoroush/code-refactoring-tool/pkg/ai/agent/mocks"
	analyzer_models "github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner/models"
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

	issues := []analyzer_models.CodeIssue{}

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
	expectedPlan := models.Plan{
		Actions: []models.PlannedAction{
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
		},
		Change: models.Change{
			Title:       "Example title for the change to let know.",
			Description: "Example description for what the change is and why is it important.",
		},
	}
	expectedPlanBytes, err := json.Marshal(expectedPlan)
	require.NoError(t, err, "MarshalJSON should not return an error")

	agnt.EXPECT().Ask(ctx, gomock.Any()).Return(string(expectedPlanBytes), nil)

	p := planner.NewAIPlanner(agnt)

	issues := []analyzer_models.CodeIssue{
		{
			Tool:          "",
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
	if !reflect.DeepEqual(plan, expectedPlan) {
		t.Errorf("expected %v, got %v", expectedPlan, plan)
	}
}
