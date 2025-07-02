// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

const (
	// CodeRefactoringAgentName is the name of the agent used for code refactoring tasks
	CodeRefactoringAgentName = "CodeRefactoringAgent"

	// CodeRefactoringAgentDescription agent description.
	CodeRefactoringAgentDescription = "Sample description"

	// CodeRefactoringAgentFoundationModel FM used for this agent.
	CodeRefactoringAgentFoundationModel = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-tg1-large"

	// CodeRefactoringAgentPrompt is the prompt used to initialize the agent for code refactoring tasks.
	CodeRefactoringAgentPrompt = "You are an agent that helps with code refactoring tasks."
)

// BedrockAgentBuilder is an implementation of AgentBuilder that uses AWS Bedrock for building agents.
type BedrockAgentBuilder struct {
	kbClient     *bedrockagent.Client
	agentRoleARN string
}

// NewBedrockAgentBuilder creates a new instance of BedrockAgentBuilder.
func NewBedrockAgentBuilder(awsConfig aws.Config, agentRoleARN string) (AgentBuilder, error) {
	return BedrockAgentBuilder{
		kbClient:     bedrockagent.NewFromConfig(awsConfig),
		agentRoleARN: agentRoleARN,
	}, nil
}

// Build implements AgentBuilder.
func (b BedrockAgentBuilder) Build(ctx context.Context, kbID string) (string, error) {
	output, err := b.kbClient.CreateAgent(ctx, &bedrockagent.CreateAgentInput{
		AgentName:            aws.String(CodeRefactoringAgentName), // TODO: How to make unique per project?
		AgentCollaboration:   types.AgentCollaborationDisabled,
		AgentResourceRoleArn: aws.String(b.agentRoleARN),
		// ClientToken *string
		// CustomOrchestration *types.CustomOrchestration
		// CustomerEncryptionKeyArn *string
		Description:     aws.String(CodeRefactoringAgentDescription),     // TODO: How to make unique per project?
		FoundationModel: aws.String(CodeRefactoringAgentFoundationModel), // TODO: Replace with actual foundation model ARN
		// GuardrailConfiguration *types.GuardrailConfiguration
		// IdleSessionTTLInSeconds *int32
		Instruction: aws.String(CodeRefactoringAgentPrompt),
		MemoryConfiguration: &types.MemoryConfiguration{ // TODO: Do we need this?
			EnabledMemoryTypes: []types.MemoryType{
				types.MemoryTypeSessionSummary,
			},
			SessionSummaryConfiguration: &types.SessionSummaryConfiguration{
				MaxRecentSessions: aws.Int32(5), // TODO: Adjust based on requirements
			},
			StorageDays: aws.Int32(30), // TODO: Adjust based on requirements
		},
		OrchestrationType: types.OrchestrationTypeDefault, // TODO: Should we use this? OrchestrationTypeCustomOrchestration,
		// PromptOverrideConfiguration *types.PromptOverrideConfiguration
		Tags: map[string]string{
			config.DefaultResourceTagKey: config.DefaultResourceTagValue,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}
	if output.Agent == nil || output.Agent.AgentId == nil {
		return "", fmt.Errorf("agent is nil in response")
	}

	_, err = b.kbClient.AssociateAgentKnowledgeBase(ctx, &bedrockagent.AssociateAgentKnowledgeBaseInput{
		AgentId:         output.Agent.AgentId,
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to associate agent with knowledge base: %w", err)
	}

	return *output.Agent.AgentId, nil
}
