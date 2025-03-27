// Package planner provides the code refactoring planner.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/kazemisoroush/code-refactor-tool/pkg/agent"
	analyzerModels "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner/models"
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
func (a *AIPlanner) Plan(ctx context.Context, _ string, issues []analyzerModels.LinterIssue) (models.Plan, error) {
	plan := models.Plan{}
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

	// Parse response to PlannedAction
	var plannedActions []models.PlannedAction
	err = json.Unmarshal([]byte(responseString), &plannedActions)
	if err != nil {
		return models.Plan{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	plan.Actions = append(plan.Actions, plannedActions...)

	return plan, nil
}

// CreatePrompt creates a prompt for the given issue
func (a *AIPlanner) CreatePrompt(issues []analyzerModels.LinterIssue) (string, error) {
	schemaExample := []models.PlannedAction{
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
	}

	// Marshal to pretty-printed JSON as schema
	schemaBytes, err := json.MarshalIndent(schemaExample, "", "  ")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf("You are an AI code refactoring agent. You will be given a linting issue and some Go code. Do not explain anything. Just return a valid JSON. Your task is to return a single JSON object matching the structure below:\n%s\n", string(schemaBytes))

	for _, issue := range issues {
		prompt = prompt + fmt.Sprintf("Lint rule violation: %s\nCode snippet: %s", issue.Message, strings.Join(issue.SourceSnippet, "\n"))
	}

	return prompt, nil
}
