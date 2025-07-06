package stack

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
)

func TestAppStack_ResourcesCreated(t *testing.T) {
	app := awscdk.NewApp(nil)

	stack := NewAppStack(app, "TestStack", &AppStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("us-east-1"),
			},
		},
	})

	template := assertions.Template_FromStack(stack.Stack, nil) // <-- Use stack.Stack

	t.Run("VPC is created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::EC2::VPC"), jsii.Number(1))
	})

	t.Run("Subnets are created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::EC2::Subnet"), jsii.Number(2))
	})

	t.Run("Security Groups are created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::EC2::SecurityGroup"), jsii.Number(1))
	})

	t.Run("S3 bucket is versioned", func(_ *testing.T) {
		template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
			"VersioningConfiguration": map[string]interface{}{
				"Status": "Enabled",
			},
		})
	})

	t.Run("IAM role for Bedrock created", func(_ *testing.T) {
		template.HasResourceProperties(jsii.String("AWS::IAM::Role"), map[string]interface{}{
			"AssumeRolePolicyDocument": map[string]interface{}{
				"Statement": []interface{}{
					map[string]interface{}{
						"Action": "sts:AssumeRole",
						"Effect": "Allow",
						"Principal": map[string]interface{}{
							"Service": "bedrock.amazonaws.com",
						},
					},
				},
			},
		})
	})

	t.Run("Secrets Manager secret created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::SecretsManager::Secret"), jsii.Number(1))
	})

	t.Run("RDS PostgreSQL Serverless cluster created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::RDS::DBInstance"), jsii.Number(1))
	})

	t.Run("ECS Cluster is created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::ECS::Cluster"), jsii.Number(1))
	})

	t.Run("ECS Task Definition is created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::ECS::TaskDefinition"), jsii.Number(1))
	})

	t.Run("IAM Role for ECS Task Execution is created", func(_ *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::IAM::Role"), jsii.Number(5))
	})
}
