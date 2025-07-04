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
	// CodeRefactoringAgentDescription agent description.
	CodeRefactoringAgentDescription = "Sample description"

	// CodeRefactoringAgentFoundationModel FM used for this agent.
	CodeRefactoringAgentFoundationModel = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-tg1-large"

	// CodeRefactoringAgentPrompt is the prompt used to initialize the agent for code refactoring tasks.
	CodeRefactoringAgentPrompt = "You are an agent that helps with code refactoring tasks."

	// DefaultAgentAliasName is the default alias name for the agent.
	DefaultAgentAliasName = "default" // TODO: Do we need to make this unique per project?

	// DefaultAgentAliasDescription is the default description for the agent alias.
	DefaultAgentAliasDescription = "Default alias for the code refactoring agent" // TODO: Do we need to make this unique per project?
)

// BedrockAgentBuilder is an implementation of AgentBuilder that uses AWS Bedrock for building agents.
type BedrockAgentBuilder struct {
	kbClient     *bedrockagent.Client
	repoPath     string
	agentRoleARN string
}

// NewBedrockAgentBuilder creates a new instance of BedrockAgentBuilder.
func NewBedrockAgentBuilder(awsConfig aws.Config, repoPath string, agentRoleARN string) AgentBuilder {
	return BedrockAgentBuilder{
		kbClient:     bedrockagent.NewFromConfig(awsConfig),
		repoPath:     repoPath,
		agentRoleARN: agentRoleARN,
	}
}

// Build implements AgentBuilder.
func (b BedrockAgentBuilder) Build(ctx context.Context, kbID string) (string, string, error) {
	createAgentOutput, err := b.kbClient.CreateAgent(ctx, &bedrockagent.CreateAgentInput{
		AgentName:            aws.String(b.repoPath),
		AgentCollaboration:   types.AgentCollaborationDisabled,
		AgentResourceRoleArn: aws.String(b.agentRoleARN),
		// ClientToken *string
		// CustomOrchestration *types.CustomOrchestration
		// CustomerEncryptionKeyArn *string
		Description:     aws.String(fmt.Sprintf("%s - %s", CodeRefactoringAgentDescription, b.repoPath)),
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
		return "", "", fmt.Errorf("failed to create agent: %w", err)
	}
	if createAgentOutput.Agent == nil || createAgentOutput.Agent.AgentId == nil {
		return "", "", fmt.Errorf("agent is nil in response")
	}

	_, err = b.kbClient.AssociateAgentKnowledgeBase(ctx, &bedrockagent.AssociateAgentKnowledgeBaseInput{
		AgentId:         createAgentOutput.Agent.AgentId,
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to associate agent with knowledge base: %w", err)
	}

	createAgentAliasOutput, err := b.kbClient.CreateAgentAlias(ctx, &bedrockagent.CreateAgentAliasInput{
		AgentId:        createAgentOutput.Agent.AgentId,
		AgentAliasName: aws.String(DefaultAgentAliasName),
		// ClientToken *string // TODO: Do we need this?
		Description: aws.String(DefaultAgentAliasDescription),
		// RoutingConfiguration []types.AgentAliasRoutingConfigurationListItem // TODO: Do we need this?
		Tags: map[string]string{
			config.DefaultResourceTagKey: config.DefaultResourceTagValue,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create agent alias: %w", err)
	}

	return *createAgentOutput.Agent.AgentId, *createAgentAliasOutput.AgentAlias.AgentAliasId, nil
}

// TearDown implements AgentBuilder.
func (b BedrockAgentBuilder) TearDown(ctx context.Context, agentID string, agentVersion string, kbID string) error {
	_, err := b.kbClient.DisassociateAgentKnowledgeBase(ctx, &bedrockagent.DisassociateAgentKnowledgeBaseInput{
		AgentId:         aws.String(agentID),
		AgentVersion:    aws.String(agentVersion),
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return fmt.Errorf("failed to disassociate agent from knowledge base: %w", err)
	}

	_, err = b.kbClient.DeleteAgent(ctx, &bedrockagent.DeleteAgentInput{
		AgentId:                aws.String(agentID),
		SkipResourceInUseCheck: true,
	})
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}
