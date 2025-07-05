// Package stack provides the CDK stack for the application infrastructure.
package stack

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const (
	// RDSAuroraDatabaseName is the name of the RDS Aurora database.
	RDSAuroraDatabaseName = "RefactorVectorDb"

	// DefaultResourceTagKey and DefaultResourceTagValue are used for tagging AWS resources
	DefaultResourceTagKey = "project"

	// DefaultResourceTagValue is the default value for the resource tag
	DefaultResourceTagValue = "CodeRefactoring"
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

// AppStackProps defines the properties for the application stack.
type AppStackProps struct {
	awscdk.StackProps
}

// AppStack is the main CDK stack for the application, containing all resources.
type AppStack struct {
	awscdk.Stack
	BedrockKnowledgeBaseRole awsiam.Role
	BedrockAgentRole         awsiam.Role
}

// NewAppStack creates a new CDK stack for the application.
func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) *AppStack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	region := *stack.Region()
	account := *stack.Account()

	// VPC for RDS and Fargate
	vpc := awsec2.NewVpc(stack, jsii.String("RefactorVpc"), &awsec2.VpcProps{
		MaxAzs: jsii.Number(2),
	})
	awscdk.Tags_Of(vpc).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// S3 Bucket
	bucketName := fmt.Sprintf("code-refactor-bucket-%s-%s", account, region)
	bucket := awss3.NewBucket(stack, jsii.String("CodeRefactorBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		Versioned:         jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})
	awscdk.Tags_Of(bucket).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

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
	awscdk.Tags_Of(secret).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// RDS Aurora Serverless v2
	rdsAuroraCluster := awsrds.NewDatabaseCluster(stack, jsii.String(RDSAuroraDatabaseName), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_15_3(),
		}),
		Vpc: vpc,
		Writer: awsrds.ClusterInstance_ServerlessV2(jsii.String("writer"), &awsrds.ServerlessV2ClusterInstanceProps{
			ScaleWithWriter: jsii.Bool(true),
		}),
		Credentials:             awsrds.Credentials_FromSecret(secret, jsii.String("postgres")),
		ServerlessV2MinCapacity: jsii.Number(0.5),
		ServerlessV2MaxCapacity: jsii.Number(1),
		RemovalPolicy:           awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(rdsAuroraCluster).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

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
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Actions: &[]*string{
							jsii.String("rds-data:ExecuteStatement"),
							jsii.String("rds-data:BatchExecuteStatement"),
							jsii.String("rds-data:BeginTransaction"),
							jsii.String("rds-data:CommitTransaction"),
							jsii.String("rds-data:RollbackTransaction"),
							jsii.String("rds-data:ExecuteSql"),
							jsii.String("rds-data:DescribeTable"),
						},
						Resources: &[]*string{
							rdsAuroraCluster.ClusterArn(),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(role).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Store the role for later use
	foundationModelResources := make([]*string, len(FoundationModels))
	for i, model := range FoundationModels {
		foundationModelResources[i] = jsii.String(fmt.Sprintf("arn:aws:bedrock:%s::foundation-model/%s", region, model))
	}

	// IAM Role for Bedrock Agent
	agentRole := awsiam.NewRole(stack, jsii.String("BedrockAgentRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("bedrock.amazonaws.com"), nil),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"BedrockAgentPolicy": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					// Model invocation permissions
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Sid:    jsii.String("AgentModelInvocationPermissions"),
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("bedrock:InvokeModel"),
						},
						Resources: &foundationModelResources,
					}),
					// Knowledge base query permissions
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Sid:    jsii.String("AgentKnowledgeBaseQuery"),
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("bedrock:Retrieve"),
							jsii.String("bedrock:RetrieveAndGenerate"),
						},
						Resources: &[]*string{
							jsii.String(fmt.Sprintf("arn:aws:bedrock:%s:%s:knowledge-base/*", region, account)),
						},
					}),
					// Prompt management console access
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Sid:    jsii.String("AgentPromptManagementConsole"),
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("bedrock:GetPrompt"),
						},
						Resources: &[]*string{
							jsii.String(fmt.Sprintf("arn:aws:bedrock:%s:%s:prompt/*", region, account)),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(agentRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// ECS Cluster and Fargate Task
	ecsCluster := awsecs.NewCluster(stack, jsii.String("RefactorCluster"), &awsecs.ClusterProps{
		Vpc: vpc,
	})
	awscdk.Tags_Of(ecsCluster).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	logGroup := awslogs.NewLogGroup(stack, jsii.String("FargateLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/ecs/code-refactor"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	taskRole := awsiam.NewRole(stack, jsii.String("RefactorTaskRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	taskDef := awsecs.NewFargateTaskDefinition(stack, jsii.String("RefactorTaskDef"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
		TaskRole:       taskRole,
	})
	awscdk.Tags_Of(taskRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Define a dedicated ECR repo for the app
	ecrRepo := awsecr.NewRepository(stack, jsii.String("RefactorEcrRepo"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("refactor-ecr-repo"),
		RemovalPolicy:  awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(ecrRepo).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Add container using an image from that ECR repo (tag must be pre-pushed)
	container := taskDef.AddContainer(jsii.String("RefactorContainer"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromEcrRepository(ecrRepo, jsii.String("latest")),
		Logging: awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("refactor"),
			LogGroup:     logGroup,
		}),
	})

	container.AddPortMappings(&awsecs.PortMapping{
		ContainerPort: jsii.Number(8080),
	})
	awscdk.Tags_Of(logGroup).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	awscdk.Tags_Of(taskDef).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	return &AppStack{
		Stack:                    stack,
		BedrockKnowledgeBaseRole: role,
		BedrockAgentRole:         agentRole,
	}
}
