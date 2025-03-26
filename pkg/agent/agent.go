package agent

import (
	"context"
)

// Agent is the interface that wraps the Ask method.
//
//go:generate mockgen -destination=./mocks/mock_agent.go -mock_names=Agent=MockAgent -package=mocks . Agent
type Agent interface {
	// Ask for prompt from the agent
	Ask(ctx context.Context, prompt string) (string, error)
}
