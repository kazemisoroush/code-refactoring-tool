// Package agent provides the interface for the agent that interacts with the user.
package agent

import (
	"context"
)

// Agent is the interface that wraps the Ask method.
//
//go:generate mockgen -destination=./mocks/mock_agent.go -mock_names=Agent=MockAgent -package=mocks . Agent
type Agent interface {
	// Ask for prompt from the agent.
	Ask(ctx context.Context, prompt string) (string, error)

	// Create an agent.
	Create(ctx context.Context) (string, error)

	// Delete agent by id.
	Delete(ctx context.Context, id string) error
}
