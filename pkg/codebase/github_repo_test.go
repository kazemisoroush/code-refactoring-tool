package codebase_test

import (
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
	"github.com/stretchr/testify/assert"
)

func TestRepository_GetPath(t *testing.T) {
	// Arrange
	gitConfig := config.GitConfig{
		CodebaseURL: "https://github.com/kazemisoroush/code-refactoring-tool",
		Token:       "<YOUR_GithubPersonalAccessToken_HERE>",
		Author:      "kazemisoroush",
		Email:       "kazemi.soroush@gmail.com",
	}
	r := codebase.NewGitHubCodebase(gitConfig)

	// Act
	path := r.GetPath()

	// Assert
	assert.Equal(t, "code-refactoring-tool", path, "GetPath should return the correct path")
}
