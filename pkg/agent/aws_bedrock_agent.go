package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type AWSBedrockAgent struct {
	Client  *bedrockruntime.Client
	ModelID string
}

func NewAWSBedrockAgent(cfg aws.Config, modelID string) Agent {
	client := bedrockruntime.NewFromConfig(cfg)
	return &AWSBedrockAgent{
		Client:  client,
		ModelID: modelID,
	}
}

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

	// Extract answer from Claude-style output
	if content, ok := resp["content"].(string); ok {
		return content, nil
	}

	// Or if it's wrapped in a list of messages
	if messages, ok := resp["messages"].([]interface{}); ok && len(messages) > 0 {
		if m, ok := messages[0].(map[string]interface{}); ok {
			return fmt.Sprintf("%v", m["content"]), nil
		}
	}

	return "", fmt.Errorf("unexpected response format: %v", resp)
}
