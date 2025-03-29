package config_test

import (
	"os"
	"testing"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Arrange: Set environment variables
	expectedRepoURL := "https://github.com/example/repo.git"
	expectedToken := "ghp_testtoken123"

	err := os.Setenv("GIT_REPO_URL", expectedRepoURL)
	require.NoError(t, err, "Setenv should not return an error")
	err = os.Setenv("GIT_TOKEN", expectedToken)
	require.NoError(t, err, "Setenv should not return an error")
	defer os.Unsetenv("GIT_REPO_URL")  //nolint:errcheck
	defer os.Unsetenv("GIT_TOKEN") //nolint:errcheck

	// Act: Load configuration
	cfg, err := config.LoadConfig()

	// Assert: Check no error and values are correctly set
	require.NoError(t, err, "LoadConfig should not return an error")
	assert.Equal(t, expectedRepoURL, cfg.Git.RepoURL, "RepoURL should match the environment variable")
	assert.Equal(t, expectedToken, cfg.Git.Token, "GitToken should match the environment variable")
}

func TestLoadConfig_MissingVariables(t *testing.T) {
	// Arrange: Clear environment variables
	err := os.Unsetenv("GIT_REPO_URL")
	require.NoError(t, err, "Unsetenv should not return an error")
	err = os.Unsetenv("GIT_TOKEN")
	require.NoError(t, err, "Unsetenv should not return an error")

	// Act: Load configuration
	_, err = config.LoadConfig()

	// Assert: Expect an error due to missing required variables
	assert.Error(t, err, "LoadConfig should return an error when required variables are missing")
}

func TestLoadConfig_InvalidGitHubURL(t *testing.T) {
	// Arrange: Set an invalid GitHub repo URL
	err := os.Setenv("GIT_REPO_URL", "https://invalid.com/repo.git")
	require.NoError(t, err, "Setenv should not return an error")
	err = os.Setenv("GIT_TOKEN", "ghp_testtoken123")
	require.NoError(t, err, "Setenv should not return an error")

	defer os.Unsetenv("GIT_REPO_URL")  //nolint:errcheck
	defer os.Unsetenv("GIT_TOKEN") //nolint:errcheck

	// Act: Attempt to load configuration
	_, err = config.LoadConfig()

	// Assert: Expect an error due to invalid URL format
	assert.Error(t, err, "LoadConfig should return an error for an invalid GitHub repository URL")
	assert.Contains(t, err.Error(), "invalid GitHub repository URL format", "Error message should indicate invalid format")
}
