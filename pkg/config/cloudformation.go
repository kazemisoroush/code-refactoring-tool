package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfn "github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

// CloudFormationClient interface for CloudFormation operations
//
//go:generate mockgen -destination=./mocks/mock_cloudformation.go -mock_names=CloudFormationClient=MockCloudFormationClient -package=mocks . CloudFormationClient
type CloudFormationClient interface {
	DescribeStacks(ctx context.Context, params *cfn.DescribeStacksInput, optFns ...func(*cfn.Options)) (*cfn.DescribeStacksOutput, error)
}

// DefaultCloudFormationClient implements CloudFormationClient using AWS SDK
type DefaultCloudFormationClient struct {
	client *cfn.Client
}

// NewCloudFormationClient creates a new CloudFormation client
func NewCloudFormationClient(cfg aws.Config) CloudFormationClient {
	return &DefaultCloudFormationClient{
		client: cfn.NewFromConfig(cfg),
	}
}

// DescribeStacks implements CloudFormationClient
func (c *DefaultCloudFormationClient) DescribeStacks(ctx context.Context, params *cfn.DescribeStacksInput, optFns ...func(*cfn.Options)) (*cfn.DescribeStacksOutput, error) {
	return c.client.DescribeStacks(ctx, params, optFns...)
}
