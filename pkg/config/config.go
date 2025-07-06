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

	// DefaultRepositoryTagKey get repository tag key per code base.
	DefaultRepositoryTagKey = "repository"

	// AWSBedrockRAGEmbeddingModel model used for Bedrock Knowledge Base embedding.
	AWSBedrockRAGEmbeddingModel = "amazon.titan-embed-text-v1"

	// AWSBedrockDataStoreEnrichmentModelARN is the ARN of the model used for context enrichment in the RAG pipeline.
	AWSBedrockDataStoreEnrichmentModelARN = "amazon.titan-text-express-v1:0"

	// AWSBedrockDataStoreParsingModelARN is the ARN of the model used for parsing in the RAG pipeline.
	AWSBedrockDataStoreParsingModelARN = "amazon.titan-text-express-v1:0"

	// AWSRegion used for aws.
	AWSRegion = "us-east-1"
)

var (
	// FoundationModels is a list of foundation models to be used in the application.
	FoundationModels = []string{
		// Anthropic Claude
		"anthropic.claude-instant-v1",
		"anthropic.claude-v2",
		"anthropic.claude-v2:1",
		"anthropic.claude-3-sonnet-20240229-v1:0",
		"anthropic.claude-3-5-sonnet-20240620-v1:0",

		// Mistral
		"mistral.mistral-7b-instruct-v0:2",
		"mistral.mistral-large-2402-v1:0",

		// Meta (Llama)
		"meta.llama2-13b-chat-v1",
		"meta.llama2-70b-chat-v1",

		// Cohere
		"cohere.command-r-v1",
		"cohere.command-r-plus-v1",

		// AI21 Labs
		"ai21.j2-mid-v1",
		"ai21.j2-ultra-v1",
		"ai21.j2-light-v1",

		// Amazon Titan (Text and Embeddings)
		"amazon.titan-text-lite-v1",
		"amazon.titan-text-express-v1",
		"amazon.titan-embed-text-v1",
	}
)

// Config represents the configuration for the application
type Config struct {
	Git            GitConfig  `envconfig:"GIT"`
	TimeoutSeconds int        `envconfig:"TIMEOUT_SECONDS" default:"180"`
	AWSConfig      aws.Config // Loaded using AWS SDK, not from env

	S3BucketName                string      `envconfig:"S3_BUCKET_NAME"`
	Account                     string      `envconfig:"ACCOUNT"`
	KnowledgeBaseServiceRoleARN string      `envconfig:"KNOWLEDGE_BASE_SERVICE_ROLE_ARN"`
	AgentServiceRoleARN         string      `envconfig:"AGENT_SERVICE_ROLE_ARN"`
	RDSPostgres                 RDSPostgres `envconfig:"RDS_POSTGRES"`
}

// RDSPostgres represents the configuration for AWS RDS Postgres
type RDSPostgres struct {
	Username             string `envconfig:"USERNAME"`
	Password             string `envconfig:"PASSWORD"`
	Host                 string `envconfig:"HOST"`
	Port                 int    `envconfig:"PORT"`
	CredentialsSecretARN string `envconfig:"CREDENTIALS_SECRET_ARN"`
	InstanceARN          string `envconfig:"INSTANCE_ARN"`
	DatabaseName         string `envconfig:"DATABASE_NAME" default:"RefactorVectorDb"`
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
	if cfg.KnowledgeBaseServiceRoleARN == "" || cfg.AgentServiceRoleARN == "" ||
		cfg.S3BucketName == "" || cfg.Account == "" || cfg.RDSPostgres.InstanceARN == "" ||
		cfg.RDSPostgres.CredentialsSecretARN == "" {
		stackName := "CodeRefactorInfra"
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
			case "BucketName":
				cfg.S3BucketName = *output.OutputValue
			case "Account":
				cfg.Account = *output.OutputValue
			case "RDSPostgresInstanceARN":
				cfg.RDSPostgres.InstanceARN = *output.OutputValue
			case "RDSPostgresCredentialsSecretARN":
				cfg.RDSPostgres.CredentialsSecretARN = *output.OutputValue
			}
		}
	}

	return cfg, nil
}
