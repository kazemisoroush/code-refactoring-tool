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
	kbRoleARN               string
	rdsCredentialsSecretARN string
	rdsAuroraClusterARN     string
	rdsAuroraDatabaseName   string
}

// NewBedrockRAG creates a new instance of BedrockRAG with the provided AWS configuration and parameters.
func NewBedrockRAG(
	awsConfig aws.Config,
	kbRoleARN string,
	rdsAurora config.RDSAurora,
) RAG {
	return &BedrockRAG{
		kbClient:                bedrockagent.NewFromConfig(awsConfig),
		kbRoleARN:               kbRoleARN,
		rdsCredentialsSecretARN: rdsAurora.CredentialsSecretARN,
		rdsAuroraClusterARN:     rdsAurora.ClusterARN,
		rdsAuroraDatabaseName:   rdsAurora.DatabaseName,
	}
}

// Create implements RAG.
func (b *BedrockRAG) Create(ctx context.Context, tableName string) (string, error) {
	// Create Bedrock Knowledge Base
	kbOutput, err := b.kbClient.CreateKnowledgeBase(ctx, &bedrockagent.CreateKnowledgeBaseInput{
		KnowledgeBaseConfiguration: &types.KnowledgeBaseConfiguration{
			Type: types.KnowledgeBaseTypeVector,
			VectorKnowledgeBaseConfiguration: &types.VectorKnowledgeBaseConfiguration{
				EmbeddingModelArn: aws.String("arn:aws:bedrock:us-west-2::embedding-model"), // TODO: make this configurable
				EmbeddingModelConfiguration: &types.EmbeddingModelConfiguration{
					BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
						Dimensions:        aws.Int32(1536),                // TODO: Example dimension size, adjust as needed
						EmbeddingDataType: types.EmbeddingDataTypeFloat32, // TODO: Adjust based on your model
					},
				},
				SupplementalDataStorageConfiguration: &types.SupplementalDataStorageConfiguration{ // TODO: Do we need this?
					StorageLocations: []types.SupplementalDataStorageLocation{
						{
							Type: types.SupplementalDataStorageLocationTypeS3,
							S3Location: &types.S3Location{
								Uri: aws.String("s3://some-bucket/table"), // TODO: make this configurable
							},
						},
					},
				},
			},
		},
		Name:        aws.String(CodeRefactoringKBName),
		RoleArn:     aws.String(b.kbRoleARN),
		Description: aws.String(CodeRefactoringKBDescription),
		StorageConfiguration: &types.StorageConfiguration{
			RdsConfiguration: &types.RdsConfiguration{
				CredentialsSecretArn: aws.String(b.rdsCredentialsSecretARN),
				DatabaseName:         aws.String(b.rdsAuroraDatabaseName),
				FieldMapping: &types.RdsFieldMapping{
					PrimaryKeyField: aws.String("id"),
					TextField:       aws.String("text"),
					VectorField:     aws.String("embedding"),
					MetadataField:   aws.String("metadata"),
				},
				ResourceArn: aws.String(b.rdsAuroraClusterARN),
				TableName:   aws.String(tableName),
			},
		},
		Tags: map[string]string{
			config.DefaultResourceTagKey: config.DefaultResourceTagValue,
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
