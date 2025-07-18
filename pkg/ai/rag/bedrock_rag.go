// Package rag provides an implementation of a Retrieval-Augmented Generation (RAG) system using AWS Bedrock.
package rag

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

const (
	// CodeRefactoringKBName is the name of the Knowledge Base used for code refactoring.
	CodeRefactoringKBName = "CodeRefactoringKnowledgeBase"

	// CodeRefactoringKBDescription is the description of the Knowledge Base used for code refactoring.
	CodeRefactoringKBDescription = "Knowledge Base for code refactoring tasks, containing vector embeddings of code snippets."
)

// BedrockRAG is an implementation of RAG that uses AWS Bedrock to create and manage a Knowledge Base for code refactoring tasks.
type BedrockRAG struct {
	kbClient                *bedrockagent.Client
	repoPath                string
	kbRoleARN               string
	rdsCredentialsSecretARN string
	RDSPostgresInstanceARN  string
	rdsPostgresDatabaseName string
}

// NewBedrockRAG creates a new instance of BedrockRAG with the provided AWS configuration and parameters.
func NewBedrockRAG(
	awsConfig aws.Config,
	repoPath string,
	kbRoleARN string,
	rdsPostgres config.RDSPostgres,
) RAG {
	return &BedrockRAG{
		kbClient:                bedrockagent.NewFromConfig(awsConfig),
		repoPath:                repoPath,
		kbRoleARN:               kbRoleARN,
		rdsCredentialsSecretARN: rdsPostgres.CredentialsSecretARN,
		RDSPostgresInstanceARN:  rdsPostgres.InstanceARN,
		rdsPostgresDatabaseName: rdsPostgres.DatabaseName,
	}
}

// Create implements RAG.
func (b *BedrockRAG) Create(ctx context.Context, tableName string) (string, error) {
	// Get the embedding model configuration
	modelConfig, exists := config.GetCurrentEmbeddingModelConfig()
	if !exists {
		return "", fmt.Errorf("unsupported embedding model: %s", config.AWSBedrockRAGEmbeddingModel)
	}

	// Build the vector knowledge base configuration
	vectorConfig := &types.VectorKnowledgeBaseConfiguration{
		EmbeddingModelArn: aws.String(modelConfig.GetEmbeddingModelARN(config.AWSRegion)),
	}

	// Add embedding configuration if the model supports it
	if modelConfig.SupportsConfiguration && modelConfig.Configuration != nil {
		vectorConfig.EmbeddingModelConfiguration = modelConfig.Configuration
	}

	// Create Bedrock Knowledge Base
	kbOutput, err := b.kbClient.CreateKnowledgeBase(ctx, &bedrockagent.CreateKnowledgeBaseInput{
		KnowledgeBaseConfiguration: &types.KnowledgeBaseConfiguration{
			Type:                             types.KnowledgeBaseTypeVector,
			VectorKnowledgeBaseConfiguration: vectorConfig,
		},
		Name:        aws.String(b.getName()),
		RoleArn:     aws.String(b.kbRoleARN),
		Description: aws.String(CodeRefactoringKBDescription),
		StorageConfiguration: &types.StorageConfiguration{
			Type: types.KnowledgeBaseStorageTypeRds,
			RdsConfiguration: &types.RdsConfiguration{
				CredentialsSecretArn: aws.String(b.rdsCredentialsSecretARN),
				DatabaseName:         aws.String(b.rdsPostgresDatabaseName),
				FieldMapping: &types.RdsFieldMapping{
					PrimaryKeyField: aws.String("id"),
					TextField:       aws.String("text"),
					VectorField:     aws.String("embedding"),
					MetadataField:   aws.String("metadata"),
				},
				ResourceArn: aws.String(b.RDSPostgresInstanceARN),
				TableName:   aws.String(tableName),
			},
		},
		Tags: map[string]string{
			config.DefaultResourceTagKey:   config.DefaultResourceTagValue,
			config.DefaultRepositoryTagKey: b.getRepositoryTag(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return *kbOutput.KnowledgeBase.KnowledgeBaseId, nil
}

// Delete implements RAG.
func (b *BedrockRAG) Delete(ctx context.Context, ragID string) error {
	// Delete the Bedrock Knowledge Base
	_, err := b.kbClient.DeleteKnowledgeBase(ctx, &bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(ragID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	return nil
}

// getName gets resource names.
func (b *BedrockRAG) getName() string {
	return b.repoPath
}

// getRepositoryTag gets repository tag name.
func (b *BedrockRAG) getRepositoryTag() string {
	return b.repoPath
}
