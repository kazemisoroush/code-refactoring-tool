package config

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the configuration for the application
type Config struct {
	RepoURL  string `envconfig:"REPO_URL" required:"true"`
	GitToken string `envconfig:"GITHUB_TOKEN" required:"true"`
}

// validateRepositoryURL ensures the RepoURL matches the expected GitHub URL pattern
func validateRepositoryURL(url string) error {
	// Regex for GitHub repo URL (HTTPS and SSH formats)
	gitHubURLRegex := `^(https:\/\/github\.com\/[\w-]+\/[\w.-]+(\.git)?|git@github\.com:[\w-]+\/[\w.-]+(\.git)?)$`
	matched, err := regexp.MatchString(gitHubURLRegex, url)
	if err != nil {
		return fmt.Errorf("failed to validate GitHub URL: %w", err)
	}
	if !matched {
		return errors.New("invalid GitHub repository URL format")
	}
	return nil
}

// LoadConfig loads and validates configuration from environment variables
func LoadConfig() (Config, error) {
	var config Config

	// Load environment variables into the struct
	err := envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate RepoURL format
	if err := validateRepositoryURL(config.RepoURL); err != nil {
		return config, fmt.Errorf("invalid GitHub repository URL: %w", err)
	}

	return config, nil
}
