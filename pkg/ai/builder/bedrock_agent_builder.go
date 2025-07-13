// Package builder contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package builder

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
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
	bedrockClient      *bedrock.Client
	bedrockAgentClient *bedrockagent.Client
	repoPath           string
	agentRoleARN       string
}

// NewBedrockAgentBuilder creates a new instance of BedrockAgentBuilder.
func NewBedrockAgentBuilder(awsConfig aws.Config, repoPath string, agentRoleARN string) AgentBuilder {
	return &BedrockAgentBuilder{
		bedrockClient:      bedrock.NewFromConfig(awsConfig),
		bedrockAgentClient: bedrockagent.NewFromConfig(awsConfig),
		repoPath:           repoPath,
		agentRoleARN:       agentRoleARN,
	}
}

// Build implements AgentBuilder.
func (b BedrockAgentBuilder) Build(ctx context.Context, kbID string) (string, string, error) {
	createAgentOutput, err := b.bedrockAgentClient.CreateAgent(ctx, &bedrockagent.CreateAgentInput{
		AgentName:            aws.String(b.repoPath),
		AgentCollaboration:   types.AgentCollaborationDisabled,
		AgentResourceRoleArn: aws.String(b.agentRoleARN),
		// ClientToken *string
		// CustomOrchestration -> OrchestrationExecutorMemberLambda
		// CustomerEncryptionKeyArn *string
		Description:     aws.String(fmt.Sprintf("%s - %s", CodeRefactoringAgentDescription, b.repoPath)),
		FoundationModel: aws.String(b.getModelARN()),
		// GuardrailConfiguration *types.GuardrailConfiguration
		// IdleSessionTTLInSeconds *int32
		Instruction: aws.String(CodeRefactoringAgentPrompt),
		MemoryConfiguration: &types.MemoryConfiguration{
			EnabledMemoryTypes: []types.MemoryType{
				types.MemoryTypeSessionSummary,
			},
			SessionSummaryConfiguration: &types.SessionSummaryConfiguration{
				MaxRecentSessions: aws.Int32(5),
			},
			StorageDays: aws.Int32(1),
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

	_, err = b.bedrockAgentClient.AssociateAgentKnowledgeBase(ctx, &bedrockagent.AssociateAgentKnowledgeBaseInput{
		AgentId:         createAgentOutput.Agent.AgentId,
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to associate agent with knowledge base: %w", err)
	}

	createAgentAliasOutput, err := b.bedrockAgentClient.CreateAgentAlias(ctx, &bedrockagent.CreateAgentAliasInput{
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

	// Create inference profile for the agent
	_, err = b.bedrockClient.CreateInferenceProfile(ctx, &bedrock.CreateInferenceProfileInput{
		InferenceProfileName: aws.String(b.getName()),
		ModelSource: &bedrocktypes.InferenceProfileModelSourceMemberCopyFrom{
			Value: b.getModelARN(),
		},
		// TODO: Do we need this?
		// ClientRequestToken *string
		Description: aws.String(b.getName()),
		Tags: []bedrocktypes.Tag{
			{
				Key:   aws.String(config.DefaultResourceTagKey),
				Value: aws.String(config.DefaultResourceTagValue),
			},
			{
				Key:   aws.String(config.DefaultRepositoryTagKey),
				Value: aws.String(b.getRepositoryTag()),
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create agent inference profile: %w", err)
	}

	return *createAgentOutput.Agent.AgentId, *createAgentAliasOutput.AgentAlias.AgentAliasId, nil
}

// TearDown implements AgentBuilder.
func (b BedrockAgentBuilder) TearDown(ctx context.Context, agentID string, agentVersion string, kbID string) error {
	_, err := b.bedrockAgentClient.DisassociateAgentKnowledgeBase(ctx, &bedrockagent.DisassociateAgentKnowledgeBaseInput{
		AgentId:         aws.String(agentID),
		AgentVersion:    aws.String(agentVersion),
		KnowledgeBaseId: aws.String(kbID),
	})
	if err != nil {
		return fmt.Errorf("failed to disassociate agent from knowledge base: %w", err)
	}

	_, err = b.bedrockAgentClient.DeleteAgent(ctx, &bedrockagent.DeleteAgentInput{
		AgentId:                aws.String(agentID),
		SkipResourceInUseCheck: true,
	})
	if err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}

	return nil
}

// getName gets resource names.
func (b BedrockAgentBuilder) getName() string {
	return b.repoPath
}

// getRepositoryTag gets repository tag name.
func (b BedrockAgentBuilder) getRepositoryTag() string {
	return b.repoPath
}

// getModelARN get the foundational model that is used by agent.
func (b BedrockAgentBuilder) getModelARN() string {
	return fmt.Sprintf("arn:aws:bedrock:%s::%s", config.AWSRegion, config.AWSBedrockAgentModel)
}
