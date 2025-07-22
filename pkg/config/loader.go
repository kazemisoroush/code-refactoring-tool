package config

import (
	"context"
	"encoding/json"
	"fmt"

	cfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Loader handles loading configuration with dependency injection
type Loader struct {
	cfnClient     CloudFormationClient
	secretsClient SecretsManagerClient
}

// NewLoader creates a new config loader with the provided clients
func NewLoader(cfnClient CloudFormationClient, secretsClient SecretsManagerClient) *Loader {
	return &Loader{
		cfnClient:     cfnClient,
		secretsClient: secretsClient,
	}
}

// LoadStackOutputs loads configuration values from CloudFormation stack outputs
func (l *Loader) LoadStackOutputs(ctx context.Context, stackName string, cfg *Config) error {
	resp, err := l.cfnClient.DescribeStacks(ctx, &cfn.DescribeStacksInput{
		StackName: &stackName,
	})
	if err != nil {
		return fmt.Errorf("failed to describe CloudFormation stack: %w", err)
	}

	for _, output := range resp.Stacks[0].Outputs {
		switch *output.OutputKey {
		case "BedrockKnowledgeBaseRoleArn":
			cfg.KnowledgeBaseServiceRoleARN = *output.OutputValue
		case "BedrockAgentRoleArn":
			cfg.AgentServiceRoleARN = *output.OutputValue
		case "RDSPostgresSchemaEnsureLambdaARN":
			cfg.RDSPostgres.SchemaEnsureLambdaARN = *output.OutputValue
		case "BucketName":
			cfg.S3BucketName = *output.OutputValue
		case "RDSPostgresInstanceARN":
			cfg.RDSPostgres.InstanceARN = *output.OutputValue
		case "RDSPostgresCredentialsSecretARN":
			cfg.RDSPostgres.CredentialsSecretARN = *output.OutputValue
		}
	}

	return nil
}

// LoadDatabaseCredentials loads database credentials from Secrets Manager
func (l *Loader) LoadDatabaseCredentials(ctx context.Context, secretARN string, cfg *Config) error {
	if secretARN == "" {
		return nil // No secret ARN provided, skip loading
	}

	result, err := l.secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve secret from Secrets Manager: %w", err)
	}

	var secret DatabaseSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Update PostgresConfig with values from Secrets Manager
	cfg.Postgres.Host = secret.Host
	cfg.Postgres.Port = secret.Port
	cfg.Postgres.Database = secret.DbName
	cfg.Postgres.Username = secret.Username
	cfg.Postgres.Password = secret.Password
	cfg.Postgres.SSLMode = "require" // Use SSL for RDS connections

	return nil
}
