// Package agent provides an implementation of an AWS Bedrock agent.
package agent

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

const (
	// SessionID can assume we have one session for now can change this later.
	SessionID = "SessionID"
)

// BedrockAgent represents an agent that interacts with AWS Bedrock.
type BedrockAgent struct {
	client       *bedrockagentruntime.Client
	AgentID      string
	AgentAliasID string
}

// NewBedrockAgent creates a new Bedrock agent with the given ID.
func NewBedrockAgent(awsConfig aws.Config, agentID string, agentAliasID string) Agent {
	return &BedrockAgent{
		client:       bedrockagentruntime.NewFromConfig(awsConfig),
		AgentID:      agentID,
		AgentAliasID: agentAliasID,
	}
}

// Ask sends a prompt to the Bedrock agent and returns the response.
func (a *BedrockAgent) Ask(ctx context.Context, prompt string) (string, error) {
	input := &bedrockagentruntime.InvokeAgentInput{
		AgentId:      aws.String(a.AgentID),
		AgentAliasId: aws.String(a.AgentAliasID),
		SessionId:    aws.String(SessionID),
		InputText:    aws.String(prompt),
	}

	resp, err := a.client.InvokeAgent(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke Bedrock agent: %w", err)
	}

	for event := range resp.GetStream().Events() {
		if content, ok := event.(*types.ResponseStreamMemberChunk); ok {
			if content.Value.Bytes != nil {
				return string(content.Value.Bytes), nil
			}
		} else {
			return "", fmt.Errorf("unexpected event type: %T", event)
		}
	}

	return "", fmt.Errorf("no valid response received from Bedrock agent")
}
