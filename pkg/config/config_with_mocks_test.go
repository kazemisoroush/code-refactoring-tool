package config_test

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config/mocks"
)

func TestLoadConfigWithMocks_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCfnClient := mocks.NewMockCloudFormationClient(ctrl)
	mockSecretsClient := mocks.NewMockSecretsManagerClient(ctrl)
	loader := config.NewLoader(mockCfnClient, mockSecretsClient)

	// Set environment variables
	expectedRepoURL := "https://github.com/example/repo.git"
	expectedToken := "ghp_testtoken123"

	err := os.Setenv("GIT_REPO_URL", expectedRepoURL)
	require.NoError(t, err)
	err = os.Setenv("GIT_TOKEN", expectedToken)
	require.NoError(t, err)
	err = os.Setenv("COGNITO_USER_POOL_ID", "us-east-1_123456789")
	require.NoError(t, err)
	err = os.Setenv("COGNITO_CLIENT_ID", "1234567890abcdef")
	require.NoError(t, err)
	err = os.Setenv("POSTGRES_PASSWORD", "testpassword123")
	require.NoError(t, err)

	defer func() {
		os.Unsetenv("GIT_REPO_URL")         //nolint:errcheck
		os.Unsetenv("GIT_TOKEN")            //nolint:errcheck
		os.Unsetenv("COGNITO_USER_POOL_ID") //nolint:errcheck
		os.Unsetenv("COGNITO_CLIENT_ID")    //nolint:errcheck
		os.Unsetenv("POSTGRES_PASSWORD")    //nolint:errcheck
	}()

	// Mock CloudFormation response
	stackOutput := &cfn.DescribeStacksOutput{
		Stacks: []cfnTypes.Stack{
			{
				Outputs: []cfnTypes.Output{
					{
						OutputKey:   aws.String("BedrockKnowledgeBaseRoleArn"),
						OutputValue: aws.String("arn:aws:iam::123456789012:role/KnowledgeBaseRole"),
					},
					{
						OutputKey:   aws.String("BedrockAgentRoleArn"),
						OutputValue: aws.String("arn:aws:iam::123456789012:role/AgentRole"),
					},
					{
						OutputKey:   aws.String("BucketName"),
						OutputValue: aws.String("my-s3-bucket"),
					},
					{
						OutputKey:   aws.String("RDSPostgresCredentialsSecretARN"),
						OutputValue: aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:rds-credentials"),
					},
				},
			},
		},
	}

	mockCfnClient.EXPECT().
		DescribeStacks(gomock.Any(), gomock.Any()).
		Return(stackOutput, nil).
		Times(1)

	// Mock Secrets Manager response
	secretValue := `{
		"username": "postgres",
		"password": "secretpassword",
		"engine": "postgres",
		"host": "mydb.cluster-xyz.us-east-1.rds.amazonaws.com",
		"port": 5432,
		"dbname": "code_refactoring_db"
	}`

	secretOutput := &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(secretValue),
	}

	mockSecretsClient.EXPECT().
		GetSecretValue(gomock.Any(), gomock.Any()).
		Return(secretOutput, nil).
		Times(1)

	// Act
	cfg, err := config.LoadConfigWithDependencies(loader)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedRepoURL, cfg.Git.RepoURL)
	assert.Equal(t, expectedToken, cfg.Git.Token)
	assert.Equal(t, "arn:aws:iam::123456789012:role/KnowledgeBaseRole", cfg.KnowledgeBaseServiceRoleARN)
	assert.Equal(t, "arn:aws:iam::123456789012:role/AgentRole", cfg.AgentServiceRoleARN)
	assert.Equal(t, "my-s3-bucket", cfg.S3BucketName)

	// Verify database credentials from Secrets Manager
	assert.Equal(t, "mydb.cluster-xyz.us-east-1.rds.amazonaws.com", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "code_refactoring_db", cfg.Postgres.Database)
	assert.Equal(t, "postgres", cfg.Postgres.Username)
	assert.Equal(t, "secretpassword", cfg.Postgres.Password)
	assert.Equal(t, "require", cfg.Postgres.SSLMode)
}

func TestLoadConfigWithMocks_NoSecretARN(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCfnClient := mocks.NewMockCloudFormationClient(ctrl)
	mockSecretsClient := mocks.NewMockSecretsManagerClient(ctrl)
	loader := config.NewLoader(mockCfnClient, mockSecretsClient)

	// Set environment variables
	err := os.Setenv("GIT_REPO_URL", "https://github.com/example/repo.git")
	require.NoError(t, err)
	err = os.Setenv("GIT_TOKEN", "ghp_testtoken123")
	require.NoError(t, err)
	err = os.Setenv("COGNITO_USER_POOL_ID", "us-east-1_123456789")
	require.NoError(t, err)
	err = os.Setenv("COGNITO_CLIENT_ID", "1234567890abcdef")
	require.NoError(t, err)
	err = os.Setenv("POSTGRES_PASSWORD", "testpassword123")
	require.NoError(t, err)

	defer func() {
		os.Unsetenv("GIT_REPO_URL")         //nolint:errcheck
		os.Unsetenv("GIT_TOKEN")            //nolint:errcheck
		os.Unsetenv("COGNITO_USER_POOL_ID") //nolint:errcheck
		os.Unsetenv("COGNITO_CLIENT_ID")    //nolint:errcheck
		os.Unsetenv("POSTGRES_PASSWORD")    //nolint:errcheck
	}()

	// Mock CloudFormation response without secret ARN
	stackOutput := &cfn.DescribeStacksOutput{
		Stacks: []cfnTypes.Stack{
			{
				Outputs: []cfnTypes.Output{
					{
						OutputKey:   aws.String("BedrockKnowledgeBaseRoleArn"),
						OutputValue: aws.String("arn:aws:iam::123456789012:role/KnowledgeBaseRole"),
					},
					{
						OutputKey:   aws.String("BedrockAgentRoleArn"),
						OutputValue: aws.String("arn:aws:iam::123456789012:role/AgentRole"),
					},
					{
						OutputKey:   aws.String("BucketName"),
						OutputValue: aws.String("my-s3-bucket"),
					},
				},
			},
		},
	}

	mockCfnClient.EXPECT().
		DescribeStacks(gomock.Any(), gomock.Any()).
		Return(stackOutput, nil).
		Times(1)

	// No Secrets Manager call expected since no secret ARN is provided

	// Act
	cfg, err := config.LoadConfigWithDependencies(loader)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "arn:aws:iam::123456789012:role/KnowledgeBaseRole", cfg.KnowledgeBaseServiceRoleARN)
	assert.Equal(t, "arn:aws:iam::123456789012:role/AgentRole", cfg.AgentServiceRoleARN)
	assert.Equal(t, "my-s3-bucket", cfg.S3BucketName)

	// Database config should use defaults/env vars since no secret was loaded
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "testpassword123", cfg.Postgres.Password)
}
