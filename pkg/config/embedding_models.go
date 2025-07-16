// Package config provides embedding model configurations for Bedrock Knowledge Base.
//
// Adding New Embedding Models:
// To add a new embedding model, add an entry to the EmbeddingModels map with:
// 1. The model ID as the key (e.g., "amazon.titan-embed-text-v3")
// 2. An EmbeddingModelConfig struct with:
//   - ModelID: the same model ID
//   - SupportsConfiguration: true if the model supports custom dimensions/config
//   - Configuration: nil if SupportsConfiguration is false, otherwise a valid config
//
// Example:
//
//	"new.embedding.model": {
//	    ModelID:               "new.embedding.model",
//	    SupportsConfiguration: true,
//	    Configuration: &types.EmbeddingModelConfiguration{
//	        BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
//	            Dimensions:        aws.Int32(1024),
//	            EmbeddingDataType: types.EmbeddingDataTypeFloat32,
//	        },
//	    },
//	},
package config

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
)

// EmbeddingModelConfig defines the configuration for an embedding model
type EmbeddingModelConfig struct {
	// ModelID is the Bedrock model identifier (e.g., "amazon.titan-embed-text-v1")
	ModelID string

	// SupportsConfiguration indicates if the model supports custom embedding configuration
	SupportsConfiguration bool

	// DefaultDimensions is the default number of dimensions for this model
	DefaultDimensions int32

	// Configuration contains the optional embedding model configuration
	Configuration *types.EmbeddingModelConfiguration
}

// GetEmbeddingModelARN returns the full ARN for the embedding model
func (e *EmbeddingModelConfig) GetEmbeddingModelARN(region string) string {
	return "arn:aws:bedrock:" + region + "::foundation-model/" + e.ModelID
}

// GetDimensions returns the dimensions for this embedding model
// Uses configured dimensions if available, otherwise returns default dimensions
func (e *EmbeddingModelConfig) GetDimensions() int32 {
	if e.SupportsConfiguration && e.Configuration != nil &&
		e.Configuration.BedrockEmbeddingModelConfiguration != nil &&
		e.Configuration.BedrockEmbeddingModelConfiguration.Dimensions != nil {
		return *e.Configuration.BedrockEmbeddingModelConfiguration.Dimensions
	}
	return e.DefaultDimensions
}

// EmbeddingModels contains all supported embedding models and their configurations
var EmbeddingModels = map[string]*EmbeddingModelConfig{ // Amazon Titan Embed Text v1 - doesn't support configurable dimensions, 1536 default
	"amazon.titan-embed-text-v1": {
		ModelID:               "amazon.titan-embed-text-v1",
		SupportsConfiguration: false,
		DefaultDimensions:     1536,
		Configuration:         nil,
	},

	// Amazon Titan Embed Text v2 - supports configurable dimensions
	"amazon.titan-embed-text-v2:0": {
		ModelID:               "amazon.titan-embed-text-v2:0",
		SupportsConfiguration: true,
		DefaultDimensions:     1024,
		Configuration: &types.EmbeddingModelConfiguration{
			BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
				Dimensions:        &[]int32{256, 512, 1024}[2], // 1024 dimensions
				EmbeddingDataType: types.EmbeddingDataTypeFloat32,
			},
		},
	},

	// Cohere Embed English v3 - supports configurable dimensions
	"cohere.embed-english-v3": {
		ModelID:               "cohere.embed-english-v3",
		SupportsConfiguration: true,
		DefaultDimensions:     1024,
		Configuration: &types.EmbeddingModelConfiguration{
			BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
				Dimensions:        &[]int32{256, 512, 1024}[2], // 1024 dimensions
				EmbeddingDataType: types.EmbeddingDataTypeFloat32,
			},
		},
	},

	// Cohere Embed Multilingual v3 - supports configurable dimensions
	"cohere.embed-multilingual-v3": {
		ModelID:               "cohere.embed-multilingual-v3",
		SupportsConfiguration: true,
		DefaultDimensions:     1024,
		Configuration: &types.EmbeddingModelConfiguration{
			BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
				Dimensions:        &[]int32{256, 512, 1024}[2], // 1024 dimensions
				EmbeddingDataType: types.EmbeddingDataTypeFloat32,
			},
		},
	},
}

// GetEmbeddingModelConfig returns the configuration for a given model ID
func GetEmbeddingModelConfig(modelID string) (*EmbeddingModelConfig, bool) {
	config, exists := EmbeddingModels[modelID]
	return config, exists
}

// GetCurrentEmbeddingModelConfig returns the configuration for the currently configured embedding model
func GetCurrentEmbeddingModelConfig() (*EmbeddingModelConfig, bool) {
	return GetEmbeddingModelConfig(AWSBedrockRAGEmbeddingModel)
}

// GetAllSupportedModelIDs returns a slice of all supported embedding model IDs
func GetAllSupportedModelIDs() []string {
	modelIDs := make([]string, 0, len(EmbeddingModels))
	for modelID := range EmbeddingModels {
		modelIDs = append(modelIDs, modelID)
	}
	return modelIDs
}

// GetModelsWithConfigSupport returns a slice of model IDs that support custom configuration
func GetModelsWithConfigSupport() []string {
	var modelIDs []string
	for modelID, config := range EmbeddingModels {
		if config.SupportsConfiguration {
			modelIDs = append(modelIDs, modelID)
		}
	}
	return modelIDs
}

// GetModelsWithoutConfigSupport returns a slice of model IDs that don't support custom configuration
func GetModelsWithoutConfigSupport() []string {
	var modelIDs []string
	for modelID, config := range EmbeddingModels {
		if !config.SupportsConfiguration {
			modelIDs = append(modelIDs, modelID)
		}
	}
	return modelIDs
}

// IsEmbeddingModelSupported checks if a given model ID is supported
func IsEmbeddingModelSupported(modelID string) bool {
	_, exists := EmbeddingModels[modelID]
	return exists
}
