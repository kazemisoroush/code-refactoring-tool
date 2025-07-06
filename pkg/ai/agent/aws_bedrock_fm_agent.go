package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// AWSBedrockAgent is an implementation of the Agent interface that uses the AWS Bedrock service.
type AWSBedrockAgent struct {
	Client  *bedrockruntime.Client
	ModelID string
}

// NewAWSBedrockFMAgent creates a new AWSBedrockAgent.
func NewAWSBedrockFMAgent(cfg aws.Config, modelID string) Agent {
	client := bedrockruntime.NewFromConfig(cfg)
	return &AWSBedrockAgent{
		Client:  client,
		ModelID: modelID,
	}
}

// Create implements Agent.
func (a *AWSBedrockAgent) Create(_ context.Context) (string, error) {
	return "", fmt.Errorf("no infra needed for foundation model invoke")
}

// Delete implements Agent.
func (a *AWSBedrockAgent) Delete(_ context.Context, _ string) error {
	return fmt.Errorf("no infra needed for foundation model invoke")
}

// Ask sends a prompt to the agent and returns the response.
func (a *AWSBedrockAgent) Ask(ctx context.Context, prompt string) (string, error) {
	payload := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 500,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	output, err := a.Client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &a.ModelID,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		return "", fmt.Errorf("bedrock invoke error: %w", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(output.Body, &resp); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	log.Println("response: ", string(output.Body))

	// Extract answer from Claude-style output
	if content, ok := resp["content"].(string); ok {
		return content, nil
	}

	// Or if it's wrapped in a list of messages
	if messages, ok := resp["content"].([]interface{}); ok && len(messages) > 0 {
		if msg, ok := messages[0].(map[string]interface{}); ok {
			if text, ok := msg["text"].(string); ok {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format: %v", resp)
}
