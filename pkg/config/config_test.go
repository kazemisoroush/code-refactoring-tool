package config_test

import (
	"os"
	"testing"

	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Arrange: Set environment variables
	expectedRepoURL := "https://github.com/example/repo.git"
	expectedToken := "ghp_testtoken123"

	os.Setenv("REPO_URL", expectedRepoURL)
	os.Setenv("GITHUB_TOKEN", expectedToken)
	defer os.Unsetenv("REPO_URL")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Act: Load configuration
	cfg, err := config.LoadConfig()

	// Assert: Check no error and values are correctly set
	require.NoError(t, err, "LoadConfig should not return an error")
	assert.Equal(t, expectedRepoURL, cfg.RepoURL, "RepoURL should match the environment variable")
	assert.Equal(t, expectedToken, cfg.GitToken, "GitToken should match the environment variable")
}

func TestLoadConfig_MissingVariables(t *testing.T) {
	// Arrange: Clear environment variables
	os.Unsetenv("REPO_URL")
	os.Unsetenv("GITHUB_TOKEN")

	// Act: Load configuration
	_, err := config.LoadConfig()

	// Assert: Expect an error due to missing required variables
	assert.Error(t, err, "LoadConfig should return an error when required variables are missing")
}
