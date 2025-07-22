package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManagerClient interface for Secrets Manager operations
//
//go:generate mockgen -destination=./mocks/mock_secretsmanager.go -mock_names=SecretsManagerClient=MockSecretsManagerClient -package=mocks . SecretsManagerClient
type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// DefaultSecretsManagerClient implements SecretsManagerClient using AWS SDK
type DefaultSecretsManagerClient struct {
	client *secretsmanager.Client
}

// NewSecretsManagerClient creates a new Secrets Manager client
func NewSecretsManagerClient(cfg aws.Config) SecretsManagerClient {
	return &DefaultSecretsManagerClient{
		client: secretsmanager.NewFromConfig(cfg),
	}
}

// GetSecretValue implements SecretsManagerClient
func (c *DefaultSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return c.client.GetSecretValue(ctx, params, optFns...)
}
