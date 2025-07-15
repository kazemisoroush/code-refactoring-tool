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
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
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
	BedrockKnowledgeBaseRole         *string
	BedrockAgentRole                 *string
	BucketName                       string
	Region                           string
	Account                          string
	RDSPostgresClusterARN            string
	RDSPostgresCredentialsSecretARN  string
	RDSPostgresSchemaEnsureLambdaARN string
}

// Resources holds the common resources that are shared across different components
type Resources struct {
	Stack   awscdk.Stack
	Vpc     awsec2.IVpc
	Account string
	Region  string
}

// NetworkingResources holds VPC and related networking components
type NetworkingResources struct {
	Vpc                    awsec2.IVpc
	SecretsManagerEndpoint awsec2.IInterfaceVpcEndpoint
}

// DatabaseResources holds RDS and related database components
type DatabaseResources struct {
	Cluster             awsrds.IDatabaseCluster
	CredentialsSecret   awssecretsmanager.ISecret
	MigrationLambda     awslambda.IFunction
	MigrationLambdaRole awsiam.Role
	MigrationLambdaSG   awsec2.ISecurityGroup
}

// BedrockResources holds Bedrock-related IAM roles and configurations
type BedrockResources struct {
	KnowledgeBaseRole awsiam.IRole
	AgentRole         awsiam.IRole
}

// ComputeResources holds ECS and Fargate resources
type ComputeResources struct {
	Cluster  awsecs.ICluster
	TaskDef  awsecs.IFargateTaskDefinition
	EcrRepo  awsecr.IRepository
	LogGroup awslogs.ILogGroup
}

// StorageResources holds S3 and other storage resources
type StorageResources struct {
	Bucket awss3.IBucket
	Name   string
}

// NewAppStack creates a new CDK stack for the application.
func NewAppStack(scope constructs.Construct, id string, props *AppStackProps) *AppStack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	resources := &Resources{
		Stack:   stack,
		Account: *stack.Account(),
		Region:  *stack.Region(),
	}

	// Create resources in logical order
	networking := createNetworkingResources(resources)
	resources.Vpc = networking.Vpc

	storage := createStorageResources(resources)
	database := createDatabaseResources(resources, networking)
	bedrock := createBedrockResources(resources, storage, database)

	// Create compute resources (ECS, Fargate, ECR)
	createComputeResources(resources, networking)

	return &AppStack{
		Stack:                            stack,
		BedrockKnowledgeBaseRole:         bedrock.KnowledgeBaseRole.RoleArn(),
		BedrockAgentRole:                 bedrock.AgentRole.RoleArn(),
		BucketName:                       storage.Name,
		Account:                          resources.Account,
		Region:                           resources.Region,
		RDSPostgresClusterARN:            *database.Cluster.ClusterArn(),
		RDSPostgresCredentialsSecretARN:  *database.CredentialsSecret.SecretArn(),
		RDSPostgresSchemaEnsureLambdaARN: *database.MigrationLambda.FunctionArn(),
	}
}

// createNetworkingResources creates VPC and related networking components
func createNetworkingResources(resources *Resources) *NetworkingResources {
	// VPC for RDS and Fargate
	vpc := awsec2.NewVpc(resources.Stack, jsii.String("RefactorVpc"), &awsec2.VpcProps{
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

	// Add VPC Endpoint for AWS Secrets Manager to allow private access from Lambda
	secretsManagerEndpoint := awsec2.NewInterfaceVpcEndpoint(resources.Stack, jsii.String("SecretsManagerVpcEndpoint"), &awsec2.InterfaceVpcEndpointProps{
		Vpc:     vpc,
		Service: awsec2.InterfaceVpcEndpointAwsService_SECRETS_MANAGER(),
		Subnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		PrivateDnsEnabled: jsii.Bool(true),
	})

	// Tag the VPC endpoint for visibility and management
	awscdk.Tags_Of(secretsManagerEndpoint).Add(
		jsii.String(DefaultResourceTagKey),
		jsii.String(DefaultResourceTagValue),
		nil,
	)

	return &NetworkingResources{
		Vpc:                    vpc,
		SecretsManagerEndpoint: secretsManagerEndpoint,
	}
}

// createStorageResources creates S3 bucket and related storage components
func createStorageResources(resources *Resources) *StorageResources {
	bucketName := fmt.Sprintf("code-refactor-bucket-%s-%s", resources.Account, resources.Region)
	bucket := awss3.NewBucket(resources.Stack, jsii.String("CodeRefactorBucket"), &awss3.BucketProps{
		BucketName:        jsii.String(bucketName),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		Versioned:         jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
	})
	awscdk.Tags_Of(bucket).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	return &StorageResources{
		Bucket: bucket,
		Name:   bucketName,
	}
}

// createDatabaseResources creates RDS cluster, secrets, and migration lambda
func createDatabaseResources(resources *Resources, networking *NetworkingResources) *DatabaseResources {
	// Secrets Manager Secret
	credentialsSecret := awssecretsmanager.NewSecret(resources.Stack, jsii.String("CodeRefactorDbSecret"), &awssecretsmanager.SecretProps{
		SecretName: jsii.String("code-refactor-db-secret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			SecretStringTemplate: jsii.String("{\"username\": \"postgres\"}"),
			GenerateStringKey:    jsii.String("password"),
			ExcludeCharacters:    jsii.String("\"@/\\"),
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(credentialsSecret).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// RDS Postgres Serverless v2
	cluster := awsrds.NewDatabaseCluster(resources.Stack, jsii.String(RDSPostgresDatabaseName), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_15_3(),
		}),
		Instances: jsii.Number(1),
		InstanceProps: &awsrds.InstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T4G, awsec2.InstanceSize_MEDIUM),
			Vpc:          networking.Vpc,
			VpcSubnets: &awsec2.SubnetSelection{
				SubnetType: awsec2.SubnetType_PUBLIC,
			},
			PubliclyAccessible: jsii.Bool(false),
		},
		DefaultDatabaseName: jsii.String(RDSPostgresDatabaseName),
		Port:                jsii.Number(5432),
		Credentials:         awsrds.Credentials_FromSecret(credentialsSecret, jsii.String("postgres")),
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,
		ClusterIdentifier:   jsii.String("code-refactor-cluster"),
	})
	awscdk.Tags_Of(cluster).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Create migration lambda and related resources
	migrationResources := createMigrationLambda(resources, networking, cluster, credentialsSecret)

	// print host and port
	fmt.Printf("RDS Postgres Cluster Endpoint: %s:%.0f\n", *cluster.ClusterEndpoint().Hostname(), *cluster.ClusterEndpoint().Port())
	fmt.Printf("RDS Postgres Credentials Secret ARN: %s\n", *credentialsSecret.SecretArn())
	fmt.Printf("RDS Postgres Migration Lambda ARN: %s\n", *migrationResources.MigrationLambda.FunctionArn())

	return &DatabaseResources{
		Cluster:             cluster,
		CredentialsSecret:   credentialsSecret,
		MigrationLambda:     migrationResources.MigrationLambda,
		MigrationLambdaRole: migrationResources.MigrationLambdaRole,
		MigrationLambdaSG:   migrationResources.MigrationLambdaSG,
	}
}

// MigrationLambdaResources holds resources specific to database migration
type MigrationLambdaResources struct {
	MigrationLambda     awslambda.IFunction
	MigrationLambdaRole awsiam.Role
	MigrationLambdaSG   awsec2.ISecurityGroup
}

// createMigrationLambda creates the database migration lambda and related resources
func createMigrationLambda(resources *Resources, networking *NetworkingResources, cluster awsrds.IDatabaseCluster, credentialsSecret awssecretsmanager.ISecret) *MigrationLambdaResources {
	// Security Group for the Migration Lambda
	migrationLambdaSG := awsec2.NewSecurityGroup(resources.Stack, jsii.String("DbMigrationLambdaSG"), &awsec2.SecurityGroupProps{
		Vpc:              networking.Vpc,
		Description:      jsii.String("Allow outbound connection to RDS Postgres for DB migrations"),
		AllowAllOutbound: jsii.Bool(true),
	})
	awscdk.Tags_Of(migrationLambdaSG).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Add inbound rule to RDS Security Group to allow connections from the Lambda SG
	cluster.Connections().AllowFrom(migrationLambdaSG, awsec2.Port_Tcp(jsii.Number(5432)), jsii.String("Allow DB migration lambda"))

	// IAM Role for the Migration Lambda
	migrationLambdaRole := awsiam.NewRole(resources.Stack, jsii.String("DbMigrationLambdaRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	awscdk.Tags_Of(migrationLambdaRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Grant permissions
	setupMigrationLambdaPermissions(migrationLambdaRole, credentialsSecret, cluster)

	lambdaPath := filepath.Join(getThisFileDir(), "../rds_schema_lambda")

	// Lambda Function for Schema Migration
	migrationLambda := awslambda.NewFunction(resources.Stack, jsii.String("DbMigrationLambda"), &awslambda.FunctionProps{
		Handler: jsii.String("handler.lambda_handler"),
		Runtime: awslambda.Runtime_PYTHON_3_12(),
		Code: awslambda.AssetCode_FromAsset(jsii.String(lambdaPath), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awslambda.Runtime_PYTHON_3_12().BundlingImage(),
				Command: jsii.Strings(
					"bash", "-c",
					"pip install -r requirements.txt -t /asset-output && cp -au . /asset-output",
				),
				User: jsii.String("root"),
			},
		}),
		Vpc: networking.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{
			migrationLambdaSG,
		},
		Environment: &map[string]*string{
			"DB_SECRET_ARN": credentialsSecret.SecretArn(),
			"DB_NAME":       jsii.String(RDSPostgresDatabaseName),
			"DB_HOST":       cluster.ClusterEndpoint().Hostname(),
			"DB_PORT":       jsii.String("5432"),
		},
		Timeout:           awscdk.Duration_Seconds(jsii.Number(10)),
		Role:              migrationLambdaRole,
		AllowPublicSubnet: jsii.Bool(true),
	})
	awscdk.Tags_Of(migrationLambda).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	return &MigrationLambdaResources{
		MigrationLambda:     migrationLambda,
		MigrationLambdaRole: migrationLambdaRole,
		MigrationLambdaSG:   migrationLambdaSG,
	}
}

// setupMigrationLambdaPermissions configures IAM permissions for the migration lambda
func setupMigrationLambdaPermissions(role awsiam.Role, credentialsSecret awssecretsmanager.ISecret, cluster awsrds.IDatabaseCluster) {
	// Grant the Lambda role permissions to write logs to CloudWatch
	role.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")))

	// For VPC access
	role.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaVPCAccessExecutionRole")))

	// Grant the Lambda role permissions to read the database secret
	credentialsSecret.GrantRead(role, nil)

	// Grant RDS Data API permissions
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
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
			cluster.ClusterArn(),
		},
	}))
}

// createBedrockResources creates Bedrock-related IAM roles
func createBedrockResources(resources *Resources, storage *StorageResources, database *DatabaseResources) *BedrockResources {
	knowledgeBaseRole := createBedrockKnowledgeBaseRole(resources, storage, database)
	agentRole := createBedrockAgentRole(resources)

	return &BedrockResources{
		KnowledgeBaseRole: knowledgeBaseRole,
		AgentRole:         agentRole,
	}
}

// createBedrockKnowledgeBaseRole creates the IAM role for Bedrock Knowledge Base
func createBedrockKnowledgeBaseRole(resources *Resources, storage *StorageResources, database *DatabaseResources) awsiam.IRole {
	role := awsiam.NewRole(resources.Stack, jsii.String("BedrockKnowledgeBaseRole"), &awsiam.RoleProps{
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
							storage.Bucket.BucketArn(),
							jsii.String(fmt.Sprintf("%s/*", *storage.Bucket.BucketArn())),
						},
					}),
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Actions: &[]*string{
							jsii.String("secretsmanager:GetSecretValue"),
						},
						Resources: &[]*string{
							database.CredentialsSecret.SecretArn(),
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
							database.Cluster.ClusterArn(),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(role).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	return role
}

// createBedrockAgentRole creates the IAM role for Bedrock Agent
func createBedrockAgentRole(resources *Resources) awsiam.IRole {
	foundationModelResources := make([]*string, len(FoundationModels))
	for i, model := range FoundationModels {
		foundationModelResources[i] = jsii.String(fmt.Sprintf("arn:aws:bedrock:%s::foundation-model/%s", resources.Region, model))
	}

	role := awsiam.NewRole(resources.Stack, jsii.String("BedrockAgentRole"), &awsiam.RoleProps{
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
							jsii.String(fmt.Sprintf("arn:aws:bedrock:%s:%s:knowledge-base/*", resources.Region, resources.Account)),
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
							jsii.String(fmt.Sprintf("arn:aws:bedrock:%s:%s:prompt/*", resources.Region, resources.Account)),
						},
					}),
				},
			}),
		},
	})
	awscdk.Tags_Of(role).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	return role
}

// createComputeResources creates ECS, Fargate, and ECR resources
func createComputeResources(resources *Resources, networking *NetworkingResources) *ComputeResources {
	// ECS Cluster
	cluster := awsecs.NewCluster(resources.Stack, jsii.String("RefactorCluster"), &awsecs.ClusterProps{
		Vpc: networking.Vpc,
	})
	awscdk.Tags_Of(cluster).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// CloudWatch Log Group
	logGroup := awslogs.NewLogGroup(resources.Stack, jsii.String("FargateLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/ecs/code-refactor"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(logGroup).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Task Role and Definition
	taskRole := awsiam.NewRole(resources.Stack, jsii.String("RefactorTaskRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	awscdk.Tags_Of(taskRole).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	taskDef := awsecs.NewFargateTaskDefinition(resources.Stack, jsii.String("RefactorTaskDef"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
		TaskRole:       taskRole,
	})
	awscdk.Tags_Of(taskDef).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// ECR Repository
	ecrRepo := awsecr.NewRepository(resources.Stack, jsii.String("RefactorEcrRepo"), &awsecr.RepositoryProps{
		RepositoryName: jsii.String("refactor-ecr-repo"),
		RemovalPolicy:  awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(ecrRepo).Add(jsii.String(DefaultResourceTagKey), jsii.String(DefaultResourceTagValue), nil)

	// Container Definition
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

	return &ComputeResources{
		Cluster:  cluster,
		TaskDef:  taskDef,
		EcrRepo:  ecrRepo,
		LogGroup: logGroup,
	}
}

func getThisFileDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get current file path")
	}
	return filepath.Dir(filename)
}
