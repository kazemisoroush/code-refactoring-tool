package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Arrange: Set environment variables
	ctx := context.Background()
	expectedRepoURL := "https://github.com/example/repo.git"
	expectedToken := "ghp_testtoken123"

	os.Setenv("REPO_URL", expectedRepoURL)
	os.Setenv("GITHUB_TOKEN", expectedToken)
	defer os.Unsetenv("REPO_URL")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Act: Load configuration
	cfg, err := config.LoadConfig(ctx)

	// Assert: Check no error and values are correctly set
	require.NoError(t, err, "LoadConfig should not return an error")
	assert.Equal(t, expectedRepoURL, cfg.RepoURL, "RepoURL should match the environment variable")
	assert.Equal(t, expectedToken, cfg.GitToken, "GitToken should match the environment variable")
}

func TestLoadConfig_MissingVariables(t *testing.T) {
	// Arrange: Clear environment variables
	ctx := context.Background()
	os.Unsetenv("REPO_URL")
	os.Unsetenv("GITHUB_TOKEN")

	// Act: Load configuration
	_, err := config.LoadConfig(ctx)

	// Assert: Expect an error due to missing required variables
	assert.Error(t, err, "LoadConfig should return an error when required variables are missing")
}

func TestLoadConfig_InvalidGitHubURL(t *testing.T) {
	// Arrange: Set an invalid GitHub repo URL
	ctx := context.Background()
	os.Setenv("REPO_URL", "https://invalid.com/repo.git")
	os.Setenv("GITHUB_TOKEN", "ghp_testtoken123")
	defer os.Unsetenv("REPO_URL")
	defer os.Unsetenv("GITHUB_TOKEN")

	// Act: Attempt to load configuration
	_, err := config.LoadConfig(ctx)

	// Assert: Expect an error due to invalid URL format
	assert.Error(t, err, "LoadConfig should return an error for an invalid GitHub repository URL")
	assert.Contains(t, err.Error(), "invalid GitHub repository URL format", "Error message should indicate invalid format")
}
