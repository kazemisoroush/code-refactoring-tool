package planner_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	agent_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/agent/mocks"
	analyzer_models "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner"
	"github.com/stretchr/testify/assert"
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
	agnt.EXPECT().Ask(ctx, gomock.Any()).Return("", nil)

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
	_, err := p.Plan(ctx, "sourcePath", issues)

	// Assert
	assert.NoError(t, err, "Plan should not return an error")
}
