package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

func TestDefaultCodebaseConfigService_Basic(t *testing.T) {
	// Basic test to ensure the service can be instantiated
	service := services.NewDefaultCodebaseConfigService(nil)
	assert.NotNil(t, service)
}

func TestProvider_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		provider models.Provider
		expected bool
	}{
		{
			name:     "valid github provider",
			provider: models.ProviderGitHub,
			expected: true,
		},
		{
			name:     "valid gitlab provider",
			provider: models.ProviderGitLab,
			expected: true,
		},
		{
			name:     "valid bitbucket provider",
			provider: models.ProviderBitbucket,
			expected: true,
		},
		{
			name:     "valid custom provider",
			provider: models.ProviderCustom,
			expected: true,
		},
		{
			name:     "invalid provider",
			provider: "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
