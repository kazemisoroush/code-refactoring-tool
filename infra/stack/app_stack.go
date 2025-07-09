// Package stack provides the CDK stack for the application infrastructure.
package stack

import (
	"fmt"
	"path/filepath"
	"runtime"

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

	// NEW IMPORT for Lambda
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	// NEW IMPORT for Custom Resources
)

const (
	// RDSPostgresDatabaseName is the name of the RDS Postgres database.
	RDSPostgresDatabaseName = "RefactorVectorDb"

	// RDSPostgresTableName table name.
	RDSPostgresTableName = "vector_store" // Define your table name here

	// DefaultResourceTagKey and DefaultResourceTagValue are used for tagging AWS resources
	DefaultResourceTagKey = "project"

	// DefaultResourceTagValue is the default value for the resource tag
	DefaultResourceTagValue = "CodeRefactoring"

	// SchemaVersion is a version string for the database schema.
	// Increment this string to trigger new migrations.
	SchemaVersion = "v1" // Change to "v2", "v3", etc., for future schema updates
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
	BedrockKnowledgeBaseRole        *string
	BedrockAgentRole                *string
	BucketName                      string
	Region                          string
	Account                         string
	RDSPostgresInstanceARN          string
	RDSPostgresCredentialsSecretARN string
}

// NewAppStack creates a new CDK stack for the application.
func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) *AppStack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	region := *stack.Region()
	account := *stack.Account()

	// VPC for RDS and Fargate
	vpc := awsec2.NewVpc(stack, jsii.String("RefactorVpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				CidrMask:   jsii.Number(24),
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
		},
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
	rdsPostgresCredentialsSecret := awssecretsmanager.NewSecret(stack, jsii.String("CodeRefactorDbSecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("code-refactor-db-secret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String("{\"username\": \"postgres\"}"),
			GenerateStringKey:    jsii.String("password"),
			ExcludeCharacters:    jsii.String("\"@/\\"),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(rdsPostgresCredentialsSecret).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// RDS Postgres Serverless v2
	rdsPostgresInstance := awsrds.NewDatabaseInstance(stack, jsii.String(RDSPostgresDatabaseName), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_17_5(),
		}),
		InstanceType:           awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		Vpc:                    vpc,
		MultiAz:                jsii.Bool(false),
		AllocatedStorage:       jsii.Number(20),
		Credentials:            awsrds.Credentials_FromSecret(rdsPostgresCredentialsSecret, jsii.String("postgres")),
		DatabaseName:           jsii.String(RDSPostgresDatabaseName),
		PubliclyAccessible:     jsii.Bool(false),
		RemovalPolicy:          awscdk.RemovalPolicy_DESTROY,
		DeleteAutomatedBackups: jsii.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
	})
	awscdk.Tags_Of(rdsPostgresInstance).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// ====================================================================
	// NEW: Database Schema Migration with Lambda Custom Resource
	// ====================================================================

	// 1. Security Group for the Migration Lambda
	// This SG will allow the Lambda to connect to the RDS instance.
	migrationLambdaSG := awsec2.NewSecurityGroup(stack, jsii.String("DbMigrationLambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Allow outbound connection to RDS Postgres for DB migrations"),
		AllowAllOutbound: jsii.Bool(true), // Allows outbound to RDS
	})
	awscdk.Tags_Of(migrationLambdaSG).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Add inbound rule to RDS Security Group to allow connections from the Lambda SG
	rdsPostgresInstance.Connections().AllowFrom(migrationLambdaSG, awsec2.Port_Tcp(jsii.Number(5432)), jsii.String("Allow DB migration lambda"))

	// 2. IAM Role for the Migration Lambda
	migrationLambdaRole := awsiam.NewRole(stack, jsii.String("DbMigrationLambdaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	awscdk.Tags_Of(migrationLambdaRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Grant the Lambda role permissions to write logs to CloudWatch
	migrationLambdaRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")))

	// For VPC access
	migrationLambdaRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")))

	// Grant the Lambda role permissions to read the database secret
	rdsPostgresCredentialsSecret.GrantRead(migrationLambdaRole, nil)

	// Grant the Lambda role permissions to connect to the RDS instance via Data API or direct
	// For direct pgx.Connect, the basic execution role might be enough for network access,
	// but explicit permissions are good practice.
	// If using RDS Data API (which pgx does not by default, but KnowledgeBase does), uncomment below:
	migrationLambdaRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
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
			rdsPostgresInstance.InstanceArn(),
		},
	}))

	lambdaPath := filepath.Join(getThisFileDir(), "../lambda")

	// 3. Lambda Function for Schema Migration
	// Ensure you have compiled your Go Lambda binary at ./lambda/migrator/main
	dbMigrationLambda := awslambda.NewFunction(stack, jsii.String("DbMigrationLambda"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2(),
		Handler: jsii.String("main"),                                         // The compiled executable name in your zip/asset
		Code:    awslambda.AssetCode_FromAsset(jsii.String(lambdaPath), nil), // Path to your compiled Go Lambda
		Vpc:     vpc,                                                         // Lambda must be in the same VPC as RDS
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC, // If your VPC only has public subnets
			// SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS, // If you have private subnets with NAT Gateway
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			migrationLambdaSG, // Attach the security group created above
		},
		Environment: &map[string]*string{
			"DB_SECRET_ARN": rdsPostgresCredentialsSecret.SecretArn(),
			"DB_HOST":       rdsPostgresInstance.DbInstanceEndpointAddress(),
			"DB_PORT":       rdsPostgresInstance.DbInstanceEndpointPort(),
			"DB_NAME":       jsii.String(RDSPostgresDatabaseName),
		},
		Timeout:           awscdk.Duration_Minutes(jsii.Number(2)), // Give it enough time
		Role:              migrationLambdaRole,
		AllowPublicSubnet: jsii.Bool(true), // Acknowledge that this Lambda is in a public subnet and won't have internet access
	})
	awscdk.Tags_Of(dbMigrationLambda).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// 4. Custom Resource to Trigger Schema Migration
	// This will invoke the Lambda function during stack deployments/updates.
	awscdk.NewCustomResource(stack, jsii.String("DbSchemaMigration"), &awscdk.CustomResourceProps{
		ServiceToken: dbMigrationLambda.FunctionArn(),
		Properties: &map[string]interface{}{
			"TableName":     jsii.String(RDSPostgresTableName),
			"SchemaVersion": jsii.String(SchemaVersion), // Change this value to trigger a new migration
		},
	})

	// ====================================================================
	// END NEW: Database Schema Migration
	// ====================================================================

	// IAM Role for Bedrock KnowledgeBase
	knowledgeBaseRole := awsiam.NewRole(stack, jsii.String("BedrockKnowledgeBaseRole"), &awsiam.RoleProps{
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
							rdsPostgresCredentialsSecret.SecretArn(),
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
							rdsPostgresInstance.InstanceArn(),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(knowledgeBaseRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

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
		Stack:                           stack,
		BedrockKnowledgeBaseRole:        knowledgeBaseRole.RoleArn(),
		BedrockAgentRole:                agentRole.RoleArn(),
		BucketName:                      bucketName,
		Account:                         account,
		Region:                          region,
		RDSPostgresInstanceARN:          *rdsPostgresInstance.InstanceArn(),
		RDSPostgresCredentialsSecretARN: *rdsPostgresCredentialsSecret.SecretArn(),
	}
}

func getThisFileDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get current file path")
	}
	return filepath.Dir(filename)
}
