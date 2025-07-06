// Package planner provides the code refactoring planner.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/agent"
	analyzerModels "github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner/models"
)

// AIPlanner is the interface that wraps the Fix method.
type AIPlanner struct {
	agent agent.Agent
}

// NewAIPlanner constructor.
func NewAIPlanner(agent agent.Agent) Planner {
	return &AIPlanner{
		agent: agent,
	}
}

// Plan fixes the code in the provided source path
func (a *AIPlanner) Plan(ctx context.Context, _ string, issues []analyzerModels.CodeIssue) (models.Plan, error) {
	if len(issues) == 0 {
		return models.Plan{}, nil
	}

	prompt, err := a.CreatePrompt(issues)
	if err != nil {
		return models.Plan{}, fmt.Errorf("failed to create prompt: %w", err)
	}
	log.Printf("prompt: %s", prompt)

	// send prompt to Bedrock
	responseString, err := a.agent.Ask(ctx, prompt)
	if err != nil {
		return models.Plan{}, fmt.Errorf("failed to ask agent: %w", err)
	}
	log.Printf("response: %s", responseString)

	// Parse response to Plan
	var plan models.Plan
	err = json.Unmarshal([]byte(responseString), &plan)
	if err != nil {
		return models.Plan{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return plan, nil
}

// CreatePrompt creates a prompt for the given issue
func (a *AIPlanner) CreatePrompt(issues []analyzerModels.CodeIssue) (string, error) {
	schemaExample := models.Plan{
		Actions: []models.PlannedAction{
			{
				FilePath: "example-file-path",
				Edits: []models.EditRegion{
					{
						StartLine: 1,
						EndLine:   1,
						Replacement: []string{
							"replacement1",
							"replacement2",
						},
					},
				},
				Reason: "example-reason",
			},
		},
		Change: models.Change{
			Title:       "Example title for the change to let know.",
			Description: "Example description for what the change is and why is it important.",
		},
	}

	// Marshal to pretty-printed JSON as schema
	schemaBytes, err := json.Marshal(schemaExample)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf("You are an AI code refactoring agent. You will be given a linting issue and some Go code. Do not explain anything. Just return a valid JSON. Your task is to return a single JSON object matching the structure below:\n%s\n", string(schemaBytes))

	issuesBytes, err := json.Marshal(issues)
	if err != nil {
		return "", err
	}

	prompt += fmt.Sprintf("Here are linting issues:\n%s\n", string(issuesBytes))

	return prompt, nil
}
