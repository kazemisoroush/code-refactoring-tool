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

	template := assertions.Template_FromStack(stack, nil)

	t.Run("S3 bucket is versioned", func(t *testing.T) {
		template.HasResourceProperties(jsii.String("AWS::S3::Bucket"), map[string]interface{}{
			"VersioningConfiguration": map[string]interface{}{
				"Status": "Enabled",
			},
		})
	})

	t.Run("IAM role for Bedrock created", func(t *testing.T) {
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

	t.Run("Secrets Manager secret created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::SecretsManager::Secret"), jsii.Number(1))
	})

	t.Run("Aurora PostgreSQL Serverless cluster created", func(t *testing.T) {
		template.ResourceCountIs(jsii.String("AWS::RDS::DBCluster"), jsii.Number(1))
	})

}
