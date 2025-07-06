package main

import (
	"infra/stack"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)

	infrastructureStack := stack.NewAppStack(app, "CodeRefactorInfra", &stack.AppStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("us-east-1"),
			},
		},
	})

	// Output BedrockKnowledgeBaseRoleArn
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("BedrockKnowledgeBaseRoleArn"), &awscdk.CfnOutputProps{
		Value: infrastructureStack.BedrockKnowledgeBaseRole,
	})

	// Output BedrockAgentRoleArn
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("BedrockAgentRoleArn"), &awscdk.CfnOutputProps{
		Value: infrastructureStack.BedrockAgentRole,
	})

	// Output BucketName
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value: &infrastructureStack.BucketName,
	})

	// Output Account
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("Account"), &awscdk.CfnOutputProps{
		Value: &infrastructureStack.Account,
	})

	// Output Region
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("Region"), &awscdk.CfnOutputProps{
		Value: &infrastructureStack.Region,
	})

	// Output RDSAuroraClusterARN
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("RDSAuroraClusterARN"), &awscdk.CfnOutputProps{
		Value: &infrastructureStack.RDSAuroraClusterARN,
	})

	// Output RDSAuroraCredentialsSecretARN
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("RDSAuroraCredentialsSecretARN"), &awscdk.CfnOutputProps{
		Value: &infrastructureStack.RDSAuroraCredentialsSecretARN,
	})

	app.Synth(nil)
}
