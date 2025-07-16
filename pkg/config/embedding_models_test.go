package config

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
)

func TestEmbeddingModelConfig_GetEmbeddingModelARN(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		region   string
		expected string
	}{
		{
			name:     "titan-embed-text-v1",
			modelID:  "amazon.titan-embed-text-v1",
			region:   "us-east-1",
			expected: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1",
		},
		{
			name:     "cohere-embed-english-v3",
			modelID:  "cohere.embed-english-v3",
			region:   "us-west-2",
			expected: "arn:aws:bedrock:us-west-2::foundation-model/cohere.embed-english-v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &EmbeddingModelConfig{ModelID: tt.modelID}
			result := config.GetEmbeddingModelARN(tt.region)
			if result != tt.expected {
				t.Errorf("GetEmbeddingModelARN() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEmbeddingModelConfig(t *testing.T) {
	tests := []struct {
		name                 string
		modelID              string
		expectExists         bool
		expectSupportsConfig bool
	}{
		{
			name:                 "existing model without config support",
			modelID:              "amazon.titan-embed-text-v1",
			expectExists:         true,
			expectSupportsConfig: false,
		},
		{
			name:                 "existing model with config support",
			modelID:              "amazon.titan-embed-text-v2:0",
			expectExists:         true,
			expectSupportsConfig: true,
		},
		{
			name:                 "non-existing model",
			modelID:              "non.existent.model",
			expectExists:         false,
			expectSupportsConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, exists := GetEmbeddingModelConfig(tt.modelID)

			if exists != tt.expectExists {
				t.Errorf("GetEmbeddingModelConfig() exists = %v, want %v", exists, tt.expectExists)
				return
			}

			if exists {
				if config.SupportsConfiguration != tt.expectSupportsConfig {
					t.Errorf("GetEmbeddingModelConfig() SupportsConfiguration = %v, want %v",
						config.SupportsConfiguration, tt.expectSupportsConfig)
				}

				if tt.expectSupportsConfig && config.Configuration == nil {
					t.Error("Expected configuration to be non-nil for model that supports configuration")
				}

				if !tt.expectSupportsConfig && config.Configuration != nil {
					t.Error("Expected configuration to be nil for model that doesn't support configuration")
				}
			}
		})
	}
}

func TestGetCurrentEmbeddingModelConfig(t *testing.T) {
	// Test that we can get the current embedding model config
	config, exists := GetCurrentEmbeddingModelConfig()

	if !exists {
		t.Errorf("GetCurrentEmbeddingModelConfig() should return a valid config for current model: %s", AWSBedrockRAGEmbeddingModel)
		return
	}

	if config.ModelID != AWSBedrockRAGEmbeddingModel {
		t.Errorf("GetCurrentEmbeddingModelConfig() ModelID = %v, want %v", config.ModelID, AWSBedrockRAGEmbeddingModel)
	}
}

func TestEmbeddingModelsConfiguration(t *testing.T) {
	// Test that all embedding models in the map have valid configurations
	for modelID, config := range EmbeddingModels {
		t.Run(modelID, func(t *testing.T) {
			if config.ModelID != modelID {
				t.Errorf("Model %s has mismatched ModelID: %s", modelID, config.ModelID)
			}

			if config.SupportsConfiguration && config.Configuration == nil {
				t.Errorf("Model %s claims to support configuration but has nil Configuration", modelID)
			}

			if !config.SupportsConfiguration && config.Configuration != nil {
				t.Errorf("Model %s claims not to support configuration but has non-nil Configuration", modelID)
			}

			// Test ARN generation
			arn := config.GetEmbeddingModelARN("us-east-1")
			expectedPrefix := "arn:aws:bedrock:us-east-1::foundation-model/"
			if len(arn) <= len(expectedPrefix) || arn[:len(expectedPrefix)] != expectedPrefix {
				t.Errorf("Model %s generated invalid ARN: %s", modelID, arn)
			}

			// If the model has configuration, validate it
			if config.Configuration != nil && config.Configuration.BedrockEmbeddingModelConfiguration != nil {
				bedrockConfig := config.Configuration.BedrockEmbeddingModelConfiguration

				// Check that dimensions are valid
				if bedrockConfig.Dimensions != nil && *bedrockConfig.Dimensions <= 0 {
					t.Errorf("Model %s has invalid dimensions: %d", modelID, *bedrockConfig.Dimensions)
				}

				// Check that embedding data type is valid
				validDataTypes := []types.EmbeddingDataType{
					types.EmbeddingDataTypeFloat32,
					types.EmbeddingDataTypeBinary,
				}

				found := false
				for _, validType := range validDataTypes {
					if bedrockConfig.EmbeddingDataType == validType {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Model %s has invalid embedding data type: %s", modelID, string(bedrockConfig.EmbeddingDataType))
				}
			}
		})
	}
}

func TestGetAllSupportedModelIDs(t *testing.T) {
	modelIDs := GetAllSupportedModelIDs()

	if len(modelIDs) == 0 {
		t.Error("Expected at least one supported model ID")
	}

	// Check that all returned IDs exist in the EmbeddingModels map
	for _, modelID := range modelIDs {
		if _, exists := EmbeddingModels[modelID]; !exists {
			t.Errorf("GetAllSupportedModelIDs() returned unknown model ID: %s", modelID)
		}
	}

	// Check that we have the expected number of models
	if len(modelIDs) != len(EmbeddingModels) {
		t.Errorf("GetAllSupportedModelIDs() returned %d models, expected %d", len(modelIDs), len(EmbeddingModels))
	}
}

func TestIsEmbeddingModelSupported(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{
			name:     "supported model",
			modelID:  "amazon.titan-embed-text-v1",
			expected: true,
		},
		{
			name:     "unsupported model",
			modelID:  "unknown.model",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmbeddingModelSupported(tt.modelID)
			if result != tt.expected {
				t.Errorf("IsEmbeddingModelSupported(%s) = %v, want %v", tt.modelID, result, tt.expected)
			}
		})
	}
}

func TestEmbeddingModelConfig_GetDimensions(t *testing.T) {
	tests := []struct {
		name               string
		config             *EmbeddingModelConfig
		expectedDimensions int32
	}{
		{
			name: "model without config support uses default dimensions",
			config: &EmbeddingModelConfig{
				ModelID:               "amazon.titan-embed-text-v1",
				SupportsConfiguration: false,
				DefaultDimensions:     1536,
				Configuration:         nil,
			},
			expectedDimensions: 1536,
		},
		{
			name: "model with config support uses configured dimensions",
			config: &EmbeddingModelConfig{
				ModelID:               "amazon.titan-embed-text-v2:0",
				SupportsConfiguration: true,
				DefaultDimensions:     1024,
				Configuration: &types.EmbeddingModelConfiguration{
					BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
						Dimensions:        aws.Int32(512),
						EmbeddingDataType: types.EmbeddingDataTypeFloat32,
					},
				},
			},
			expectedDimensions: 512,
		},
		{
			name: "model with config support but no dimensions config uses default",
			config: &EmbeddingModelConfig{
				ModelID:               "test.model",
				SupportsConfiguration: true,
				DefaultDimensions:     768,
				Configuration: &types.EmbeddingModelConfiguration{
					BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
						EmbeddingDataType: types.EmbeddingDataTypeFloat32,
						// No Dimensions specified
					},
				},
			},
			expectedDimensions: 768,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDimensions()
			if result != tt.expectedDimensions {
				t.Errorf("GetDimensions() = %v, want %v", result, tt.expectedDimensions)
			}
		})
	}
}
