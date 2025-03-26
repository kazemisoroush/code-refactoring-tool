package fixer

import (
	"context"
	"fmt"
	"strings"

	"github.com/kazemisoroush/code-refactor-tool/pkg/agent"
	analyzerModels "github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

// AIFixer is the interface that wraps the Fix method.
type AIFixer struct {
	agent agent.Agent
}

// NewAIFixer constructor.
func NewAIFixer(agent agent.Agent) Fixer {
	return &AIFixer{
		agent: agent,
	}
}

// Fix fixes the code in the provided source path
func (a *AIFixer) Fix(ctx context.Context, _ string, issues []analyzerModels.LinterIssue) error {
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
			return fmt.Errorf("failed to ask agent: %w", err)
		}

		// TODO: replace the original line with the corrected one
	}
	
	return nil
}
