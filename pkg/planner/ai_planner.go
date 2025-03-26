package planner

import (
	"context"
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
	for _, issue := range issues {
		if len(issue.SourceSnippet) == 0 {
			continue
		}

		prompt := fmt.Sprintf(`You are an AI code refactoring agent.
Your task is to fix a linting issue in the following Go code snippet. Do not change code behavior.

Lint rule violation: %s

Original code:
%s

Please provide only the corrected version of this line.`,
			issue.Message,
			strings.Join(issue.SourceSnippet, "\n"),
		)

		// send prompt to Bedrock
		_, err := a.agent.Ask(ctx, prompt)
		if err != nil {
			return models.Plan{}, fmt.Errorf("failed to ask agent: %w", err)
		}

		// TODO: replace the original line with the corrected one
	}

	return models.Plan{}, nil
}
