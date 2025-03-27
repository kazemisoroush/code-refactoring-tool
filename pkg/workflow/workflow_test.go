package workflow_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	analyzer_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/mocks"
	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	ptchr_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/patcher/mocks"
	planner_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/planner/mocks"
	planner_model "github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
	repo_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/repository/mocks"
	"github.com/kazemisoroush/code-refactor-tool/pkg/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_Run(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithTimeout(context.Background(), 30)
	defer cancel()

	err := os.Setenv("REPO_URL", "https://github.com/kazemisoroush/code-refactor-tool.git")
	require.NoError(t, err, "Setenv should not return an error")
	err = os.Setenv("GITHUB_TOKEN", "some_github_token")
	require.NoError(t, err, "Setenv should not return an error")
	cfg, err := config.LoadConfig()
	require.NoError(t, err, "LoadConfig should not return an error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	analysisResult := models.AnalysisResult{}
	issues := []models.LinterIssue{
		{
			LinterName:    "test-linter",
			RuleID:        "test-rule",
			Message:       "test-message",
			FilePath:      "test-file-path",
			Line:          1,
			Column:        1,
			SourceSnippet: []string{"test-snippet"},
			Suggestions:   []string{"test-suggestion"},
		},
	}
	repoPath := "code-refactor-tool"
	plan := planner_model.Plan{}

	anlz := analyzer_mocks.NewMockAnalyzer(ctrl)
	anlz.EXPECT().AnalyzeCode(repoPath).Return(analysisResult, nil)
	anlz.EXPECT().ExtractIssues(analysisResult).Return(issues, nil)

	repo := repo_mocks.NewMockRepository(ctrl)
	repo.EXPECT().GetPath().Return(repoPath)
	repo.EXPECT().Clone().Return(nil)
	repo.EXPECT().CheckoutBranch(gomock.Any()).Return(nil)
	repo.EXPECT().Commit(gomock.Any()).Return(nil)
	repo.EXPECT().Push().Return(nil)
	repo.EXPECT().CreatePR(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("output", nil)
	repo.EXPECT().Cleanup().Return(nil)

	plnr := planner_mocks.NewMockPlanner(ctrl)
	plnr.EXPECT().Plan(ctx, repoPath, issues).Return(plan, nil)

	ptchr := ptchr_mocks.NewMockPatcher(ctrl)
	ptchr.EXPECT().Patch(repoPath, plan).Return(nil)

	wf, err := workflow.NewWorkflow(cfg, anlz, repo, plnr, ptchr)
	require.NoError(t, err, "NewWorkflow should not return an error")

	// Act
	err = wf.Run(ctx)

	// Assert
	assert.NoError(t, err, "Run should not return an error")
}
