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
	// Create Bedrock Knowledge Base
	kbOutput, err := b.kbClient.CreateKnowledgeBase(ctx, &bedrockagent.CreateKnowledgeBaseInput{
		KnowledgeBaseConfiguration: &types.KnowledgeBaseConfiguration{
			Type: types.KnowledgeBaseTypeVector, VectorKnowledgeBaseConfiguration: &types.VectorKnowledgeBaseConfiguration{
				EmbeddingModelArn: aws.String(fmt.Sprintf("arn:aws:bedrock:%s::foundation-model/%s", config.AWSRegion, config.AWSBedrockRAGEmbeddingModel)),
				EmbeddingModelConfiguration: &types.EmbeddingModelConfiguration{
					BedrockEmbeddingModelConfiguration: &types.BedrockEmbeddingModelConfiguration{
						// High accuracy needed
						// Used by models like text-embedding-3-large from OpenAI or CodeT5+
						Dimensions:        aws.Int32(1536),
						EmbeddingDataType: types.EmbeddingDataTypeFloat32,
					},
				},
			},
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
				// arn:aws:rds:us-east-1:698315877107:db:coderefactorinfra-refactorvectordb152eceac-pnzbicpefp4r
				// arn:aws(-cn|-us-gov|-eusc|-iso(-[b-f])?)?:rds:[a-zA-Z0-9-]*:[0-9]{12}:cluster:[a-zA-Z0-9-]{1,63}
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
