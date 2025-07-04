// Package config provides the configuration for the application.
package config

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	cfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/kelseyhightower/envconfig"
)

const (
	// DefaultResourceTagKey and DefaultResourceTagValue are used for tagging AWS resources
	DefaultResourceTagKey = "project"

	// DefaultResourceTagValue is the default value for the resource tag
	DefaultResourceTagValue = "CodeRefactoring"
)

// Config represents the configuration for the application
type Config struct {
	Git            GitConfig  `envconfig:"GIT"`
	TimeoutSeconds int        `envconfig:"TIMEOUT_SECONDS" default:"180"`
	AWSConfig      aws.Config // Loaded using AWS SDK, not from env

	S3BucketName                string    `envconfig:"S3_BUCKET_NAME" required:"true"`
	KnowledgeBaseServiceRoleARN string    `envconfig:"KNOWLEDGE_BASE_SERVICE_ROLE_ARN" required:"true"`
	AgentServiceRoleARN         string    `envconfig:"AGENT_SERVICE_ROLE_ARN" required:"true"`
	RDSAurora                   RDSAurora `envconfig:"RDS_AURORA" required:"true"`
}

// RDSAurora represents the configuration for AWS RDS Aurora
type RDSAurora struct {
	CredentialsSecretARN string `envconfig:"RDS_CREDENTIALS_SECRET_ARN" required:"true"`
	ClusterARN           string `envconfig:"RDS_AURORA_CLUSTER_ARN" required:"true"`
	DatabaseName         string `envconfig:"RDS_AURORA_DATABASE_NAME" default:"RefactorVectorDb"`
}

// GitConfig represents the Git configuration
type GitConfig struct {
	RepoURL string `envconfig:"REPO_URL" required:"true"`
	Token   string `envconfig:"TOKEN" required:"true"`
	Author  string `envconfig:"AUTHOR" default:"CodeRefactorBot"`
	Email   string `envconfig:"EMAIL" default:"bot@example.com"`
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

// LoadConfig loads and validates configuration from environment variables and AWS
func LoadConfig() (Config, error) {
	var cfg Config

	// Load env vars
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate RepoURL
	if err := validateRepositoryURL(cfg.Git.RepoURL); err != nil {
		return cfg, fmt.Errorf("invalid GitHub repository URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	cfg.AWSConfig = awsCfg

	// Populate BedrockKnowledgeBaseRoleARN and AgentServiceRoleARN from CloudFormation outputs if not set
	if cfg.KnowledgeBaseServiceRoleARN == "" || cfg.AgentServiceRoleARN == "" {
		stackName := "CodeRefactorInfra" // Change if your stack name is different
		cfnClient := cfn.NewFromConfig(awsCfg)
		resp, err := cfnClient.DescribeStacks(ctx, &cfn.DescribeStacksInput{
			StackName: &stackName,
		})
		if err != nil {
			return cfg, fmt.Errorf("failed to describe CloudFormation stack: %w", err)
		}
		for _, output := range resp.Stacks[0].Outputs {
			switch *output.OutputKey {
			case "BedrockKnowledgeBaseRoleArn":
				cfg.KnowledgeBaseServiceRoleARN = *output.OutputValue
			case "BedrockAgentRoleArn":
				cfg.AgentServiceRoleARN = *output.OutputValue
			}
		}
	}

	return cfg, nil
}
