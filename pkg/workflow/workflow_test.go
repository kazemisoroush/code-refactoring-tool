package workflow_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	analyzer_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/mocks"
	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	repo_mocks "github.com/kazemisoroush/code-refactor-tool/pkg/repository/mocks"
	"github.com/kazemisoroush/code-refactor-tool/pkg/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_Run(t *testing.T) {
	// Arrange
	os.Setenv("REPO_URL", "https://github.com/kazemisoroush/code-refactor-tool.git")
	os.Setenv("GITHUB_TOKEN", "some_github_token")
	cfg, err := config.LoadConfig()
	require.NoError(t, err, "LoadConfig should not return an error")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	analysisResult := models.AnalysisResult{}
	codeMetrics := models.CodeMetrics{}
	report := models.Report{}
	repoPath := "code-refactor-tool"

	a := analyzer_mocks.NewMockAnalyzer(ctrl)
	a.EXPECT().AnalyzeCode(repoPath).Return(analysisResult, nil)
	a.EXPECT().ExtractMetrics(analysisResult).Return(codeMetrics, nil)
	a.EXPECT().GenerateReport(codeMetrics).Return(report)

	r := repo_mocks.NewMockRepository(ctrl)
	r.EXPECT().GetPath().Return(repoPath)
	r.EXPECT().Clone().Return(nil)

	wf, err := workflow.NewWorkflow(cfg, a, r)
	require.NoError(t, err, "NewWorkflow should not return an error")

	// Act
	err = wf.Run()

	// Assert
	assert.NoError(t, err, "Run should not return an error")
}
