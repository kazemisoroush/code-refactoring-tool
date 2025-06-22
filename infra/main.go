package main

import (
	"infra/stack"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	app := awscdk.NewApp(nil)

	stack.NewAppStack(app, "CodeRefactorInfra", &stack.AppStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Region: jsii.String("ap-southeast-2"),
			},
		},
	})

	app.Synth(nil)
}
