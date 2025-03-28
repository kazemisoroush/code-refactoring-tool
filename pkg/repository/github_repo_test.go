package repository_test

import (
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepository_GetPath(t *testing.T) {
	// Arrange
	repositoryURL := "https://github.com/kazemisoroush/code-refactoring-tool"
	githubToken := "some_github_token"
	r := repository.NewGitHubRepo(repositoryURL, githubToken)

	// Act
	path := r.GetPath()

	// Assert
	assert.Equal(t, "code-refactoring-tool", path, "GetPath should return the correct path")
}
