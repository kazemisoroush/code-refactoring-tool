// Package planner provides the code refactoring planner.
package planner

import (
	"context"
	"encoding/json"
	"fmt"
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
	for _, issue := range issues {
		if len(issue.SourceSnippet) == 0 {
			continue
		}

		prompt, err := a.CreatePrompt(issue)
		if err != nil {
			return plan, fmt.Errorf("failed to create prompt: %w", err)
		}

		// send prompt to Bedrock
		responseString, err := a.agent.Ask(ctx, prompt)
		if err != nil {
			return plan, fmt.Errorf("failed to ask agent: %w", err)
		}

		// Parse response to PlannedAction
		var plannedActions []models.PlannedAction
		err = json.Unmarshal([]byte(responseString), &plannedActions)
		if err != nil {
			return plan, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		plan.Actions = append(plan.Actions, plannedActions...)
	}

	return plan, nil
}

// CreatePrompt creates a prompt for the given issue
func (a *AIPlanner) CreatePrompt(issue analyzerModels.LinterIssue) (string, error) {
	schemaExample := []models.PlannedAction{}

	// Marshal to pretty-printed JSON as schema
	schemaBytes, err := json.MarshalIndent(schemaExample, "", "  ")
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`You are an AI code refactoring agent.
You will be given a linting issue and some Go code. Your task is to return a single JSON object matching the structure below:

%s

Do not explain anything. Just return a valid JSON.

Lint rule violation: %s

Code snippet:
%s
`, string(schemaBytes), issue.Message, strings.Join(issue.SourceSnippet, "\n"))

	return prompt, nil
}
