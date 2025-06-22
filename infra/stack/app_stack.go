package stack

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AppStackProps struct {
	awscdk.StackProps
}

func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	projectTag := jsii.String("code-refactoring-tool")
	region := *stack.Region()
	account := *stack.Account()

	// VPC for RDS
	vpc := awsec2.NewVpc(stack, jsii.String("RefactorVpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})
	awscdk.Tags_Of(vpc).Add(jsii.String("Project"), projectTag, nil)

	// S3 Bucket
	bucketName := fmt.Sprintf("code-refactor-bucket-%s-%s", account, region)
	bucket := awss3.NewBucket(stack, jsii.String("CodeRefactorBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		Versioned:         jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})
	awscdk.Tags_Of(bucket).Add(jsii.String("Project"), projectTag, nil)

	// Secrets Manager Secret
	secret := awssecretsmanager.NewSecret(stack, jsii.String("CodeRefactorDbSecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("code-refactor-db-secret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String("{\"username\": \"postgres\"}"),
			GenerateStringKey:    jsii.String("password"),
			ExcludeCharacters:    jsii.String("\"@/\\"),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(secret).Add(jsii.String("Project"), projectTag, nil)

	// RDS Aurora Serverless v2
	cluster := awsrds.NewDatabaseCluster(stack, jsii.String("RefactorVectorDb"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_15_3(),
		}),
		Credentials: awsrds.Credentials_FromSecret(secret, jsii.String("postgres")),
		InstanceProps: &awsrds.InstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MEDIUM),
			Vpc:          vpc,
		},
		ServerlessV2MinCapacity: jsii.Number(0.5),
		ServerlessV2MaxCapacity: jsii.Number(2),
		RemovalPolicy:           awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(cluster).Add(jsii.String("Project"), projectTag, nil)

	// IAM Role for Bedrock KnowledgeBase
	role := awsiam.NewRole(stack, jsii.String("BedrockKnowledgeBaseRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("bedrock.amazonaws.com"), nil),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"BedrockKbPolicy": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Actions: &[]*string{
							jsii.String("s3:GetObject"),
							jsii.String("s3:ListBucket"),
						},
						Resources: &[]*string{
							bucket.BucketArn(),
							jsii.String(fmt.Sprintf("%s/*", *bucket.BucketArn())),
						},
					}),
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Actions: &[]*string{
							jsii.String("secretsmanager:GetSecretValue"),
						},
						Resources: &[]*string{
							secret.SecretArn(),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(role).Add(jsii.String("Project"), projectTag, nil)

	return stack
}
