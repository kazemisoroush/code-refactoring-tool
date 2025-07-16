# Embedding Model Configuration

This document explains how to configure and extend embedding models for the Bedrock Knowledge Base.

## Overview

The embedding model configuration system allows you to:
- Define which embedding models are supported
- Specify whether a model supports custom configuration (dimensions, data types)
- Provide default configurations for models that support them
- Easily add new models without changing business logic

## Current Supported Models

### Models without configuration support:
- `amazon.titan-embed-text-v1` - Basic Titan embedding model

### Models with configuration support:
- `amazon.titan-embed-text-v2:0` - Supports custom dimensions (1024 default)
- `cohere.embed-english-v3` - Supports custom dimensions (1024 default)
- `cohere.embed-multilingual-v3` - Supports custom dimensions (1024 default)

## Adding New Models

To add a new embedding model, edit `pkg/config/embedding_models.go` and add an entry to the `EmbeddingModels` map:

```go
// For a model that doesn't support configuration
"new.basic.model": {
    ModelID:               "new.basic.model",
    SupportsConfiguration: false,
    Configuration:         nil,
},

// For a model that supports configuration
"new.advanced.model": {
    ModelID:               "new.advanced.model",
    SupportsConfiguration: true,
    Configuration: &types.EmbeddingModelConfiguration{
        BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
            Dimensions:        aws.Int32(768), // or 256, 512, 1024, etc.
            EmbeddingDataType: types.EmbeddingDataTypeFloat32,
        },
    },
},
```

## Usage

### Getting the current model configuration
```go
config, exists := config.GetCurrentEmbeddingModelConfig()
if !exists {
    return fmt.Errorf("unsupported embedding model: %s", config.AWSBedrockRAGEmbeddingModel)
}
```

### Building ARN for a model
```go
arn := config.GetEmbeddingModelARN("us-east-1")
// Returns: "arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-embed-text-v1"
```

### Checking if a model supports configuration
```go
if config.SupportsConfiguration && config.Configuration != nil {
    vectorConfig.EmbeddingModelConfiguration = config.Configuration
}
```

### Helper functions
```go
// Get all supported model IDs
allModels := config.GetAllSupportedModelIDs()

// Get models that support configuration
advancedModels := config.GetModelsWithConfigSupport()

// Get models that don't support configuration
basicModels := config.GetModelsWithoutConfigSupport()

// Check if a model is supported
if config.IsEmbeddingModelSupported("amazon.titan-embed-text-v1") {
    // Model is supported
}
```

## Configuration Details

### Supported Dimensions
Different models support different dimension ranges:
- Titan models: typically 1024 dimensions
- Cohere models: 256, 512, or 1024 dimensions
- Check AWS Bedrock documentation for specific model limits

### Supported Data Types
- `types.EmbeddingDataTypeFloat32` - Most common, good balance of precision and size
- `types.EmbeddingDataTypeBinary` - More compact but may reduce precision

## Testing

The system includes comprehensive tests to ensure:
- All models have valid configurations
- ARN generation works correctly
- Helper functions return expected results
- Configuration validation works properly

Run tests with:
```bash
go test -v ./pkg/config/
```

## Changing the Default Model

To change which embedding model is used by default, update the `AWSBedrockRAGEmbeddingModel` constant in `pkg/config/config.go`:

```go
const AWSBedrockRAGEmbeddingModel = "amazon.titan-embed-text-v2:0"
```

Make sure the model you choose is defined in the `EmbeddingModels` map.
