package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config represents the configuration for the application
type Config struct {
	RepoURL  string `envconfig:"REPO_URL" required:"true"`
	GitToken string `envconfig:"GITHUB_TOKEN" required:"true"`
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() (Config, error) {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
