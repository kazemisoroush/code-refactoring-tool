package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentStatus_Values(t *testing.T) {
	assert.Equal(t, AgentStatus("pending"), AgentStatusPending)
	assert.Equal(t, AgentStatus("initializing"), AgentStatusInitializing)
	assert.Equal(t, AgentStatus("ready"), AgentStatusReady)
	assert.Equal(t, AgentStatus("failed"), AgentStatusFailed)
}
