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
		Value: infrastructureStack.BedrockKnowledgeBaseRole.RoleArn(),
	})

	// Output BedrockAgentRoleArn
	awscdk.NewCfnOutput(infrastructureStack.Stack, jsii.String("BedrockAgentRoleArn"), &awscdk.CfnOutputProps{
		Value: infrastructureStack.BedrockAgentRole.RoleArn(),
	})

	app.Synth(nil)
}
