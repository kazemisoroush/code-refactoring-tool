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
	"github.com/kelseyhightower/envconfig"
)

// Config represents the configuration for the application
type Config struct {
	Git            GitConfig      `envconfig:"GIT"`
	TimeoutSeconds int            `envconfig:"TIMEOUT_SECONDS" default:"180"`
	LogLevel       string         `envconfig:"LOG_LEVEL" default:"info"`
	AWSConfig      aws.Config     // Loaded using AWS SDK, not from env
	Cognito        CognitoConfig  `envconfig:"COGNITO"`
	Metrics        MetricsConfig  `envconfig:"METRICS"`
	Postgres       PostgresConfig `envconfig:"POSTGRES"`

	// AI configuration (organized by provider)
	AI AIConfig `envconfig:"AI"`
}

// LocalAIConfig represents the configuration for local AI services
type LocalAIConfig struct {
	Enabled        bool   `envconfig:"ENABLED" default:"false"`
	OllamaURL      string `envconfig:"OLLAMA_URL" default:"http://localhost:11434"`
	Model          string `envconfig:"MODEL" default:"codellama:7b-instruct"`
	ChromaURL      string `envconfig:"CHROMA_URL" default:"http://localhost:8000"`
	EmbeddingModel string `envconfig:"EMBEDDING_MODEL" default:"all-MiniLM-L6-v2"`
}

// BedrockAIConfig represents the configuration for AWS Bedrock AI services
type BedrockAIConfig struct {
	Region                      string      `envconfig:"REGION" default:"us-east-1"`
	KnowledgeBaseServiceRoleARN string      `envconfig:"KNOWLEDGE_BASE_SERVICE_ROLE_ARN"`
	AgentServiceRoleARN         string      `envconfig:"AGENT_SERVICE_ROLE_ARN"`
	FoundationModel             string      `envconfig:"FOUNDATION_MODEL" default:"amazon.titan-tg1-large"`
	S3BucketName                string      `envconfig:"S3_BUCKET_NAME"`
	RDSPostgres                 RDSPostgres `envconfig:"RDS_POSTGRES"`
}

// AIConfig represents the overall AI configuration with provider-specific settings
type AIConfig struct {
	// Provider selection (can be overridden per request)
	DefaultProvider string `envconfig:"DEFAULT_PROVIDER" default:"bedrock"`

	// Provider-specific configurations
	Local   LocalAIConfig   `envconfig:"LOCAL"`
	Bedrock BedrockAIConfig `envconfig:"BEDROCK"`
}

// RDSPostgres represents the configuration for AWS RDS Postgres
type RDSPostgres struct {
	CredentialsSecretARN  string `envconfig:"CREDENTIALS_SECRET_ARN"`
	SchemaEnsureLambdaARN string `envconfig:"RDS_POSTGRES_SCHEMA_ENSURE_LAMBDA_ARN"`
	InstanceARN           string `envconfig:"INSTANCE_ARN"`
	DatabaseName          string `envconfig:"DATABASE_NAME" default:"code_refactoring_db"`
}

// CognitoConfig represents the configuration for AWS Cognito authentication
type CognitoConfig struct {
	UserPoolID string `envconfig:"USER_POOL_ID" required:"true"`
	ClientID   string `envconfig:"CLIENT_ID" required:"true"`
	Region     string `envconfig:"REGION" default:"us-east-1"`
}

// MetricsConfig represents the configuration for metrics collection
type MetricsConfig struct {
	Namespace   string `envconfig:"NAMESPACE" default:"CodeRefactorTool/API"`
	Region      string `envconfig:"REGION" default:"us-east-1"`
	ServiceName string `envconfig:"SERVICE_NAME" default:"code-refactor-api"`
	Enabled     bool   `envconfig:"ENABLED" default:"true"`
}

// PostgresConfig represents the configuration for PostgreSQL connection
type PostgresConfig struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" default:"5432"`
	Database string `envconfig:"DATABASE" default:"code_refactoring_db"`
	Username string `envconfig:"USERNAME" default:"postgres"`
	Password string `envconfig:"PASSWORD"`
	SSLMode  string `envconfig:"SSL_MODE" default:"disable"`
}

// DatabaseSecret represents the structure of the secret stored in AWS Secrets Manager
type DatabaseSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DbName   string `json:"dbname"`
}

// GitConfig represents the Git configuration
type GitConfig struct {
	CodebaseURL string `envconfig:"CODEBASE_URL"` // Per-request, not required at startup
	Token       string `envconfig:"TOKEN" required:"true"`
	Author      string `envconfig:"AUTHOR" default:"CodeRefactorBot"`
	Email       string `envconfig:"EMAIL" default:"bot@example.com"`
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
	return LoadConfigWithDependencies(nil)
}

// LoadConfigWithDependencies loads configuration with optional dependency injection for testing
func LoadConfigWithDependencies(loader *Loader) (Config, error) {
	var cfg Config

	// Load env vars
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Validate RepoURL if provided (it's optional at startup, provided per-request)
	if cfg.Git.CodebaseURL != "" {
		if err := validateRepositoryURL(cfg.Git.CodebaseURL); err != nil {
			return cfg, fmt.Errorf("invalid GitHub repository URL: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return cfg, fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	cfg.AWSConfig = awsCfg

	// Use provided loader or create default one
	if loader == nil {
		cfnClient := NewCloudFormationClient(awsCfg)
		secretsClient := NewSecretsManagerClient(awsCfg)
		loader = NewLoader(cfnClient, secretsClient)
	}

	// Only load AWS resources if not using local AI
	if !cfg.AI.Local.Enabled {
		// Load values from CloudFormation stack if not already set
		if cfg.AI.Bedrock.KnowledgeBaseServiceRoleARN == "" || cfg.AI.Bedrock.AgentServiceRoleARN == "" ||
			cfg.AI.Bedrock.S3BucketName == "" || cfg.AI.Bedrock.RDSPostgres.InstanceARN == "" || cfg.AI.Bedrock.RDSPostgres.SchemaEnsureLambdaARN == "" ||
			cfg.AI.Bedrock.RDSPostgres.CredentialsSecretARN == "" {
			if err := loader.LoadStackOutputs(ctx, "CodeRefactorInfra", &cfg); err != nil {
				return cfg, fmt.Errorf("failed to load stack outputs: %w", err)
			}
		}

		// Load database credentials from Secrets Manager if secret ARN is available
		if err := loader.LoadDatabaseCredentials(ctx, cfg.AI.Bedrock.RDSPostgres.CredentialsSecretARN, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to load database credentials: %w", err)
		}
	}

	return cfg, nil
}
