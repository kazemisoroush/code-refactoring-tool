// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

const (
	// ProviderName is the name of the provider used for building the RAG pipeline.
	ProviderName = "bedrock"
	// CodeRefactoringKBName is the name of the Knowledge Base used for code refactoring.
	CodeRefactoringKBName = "CodeRefactoringKnowledgeBase"
	// CodeRefactoringKBDescription is the description of the Knowledge Base used for code refactoring.
	CodeRefactoringKBDescription = "Knowledge Base for code refactoring tasks, containing vector embeddings of code snippets."
	// RDSAuroraDatabaseName is the name of the RDS Aurora database.
	RDSAuroraDatabaseName = "RefactorVectorDb"
)

// BedrockRAGBuilder is an implementation of RAGBuilder that uses AWS Bedrock for building the RAG pipeline.
type BedrockRAGBuilder struct {
	s3Client                *s3.Client
	kbClient                *bedrockagent.Client
	rdsClient               *rdsdata.Client
	repository              repository.Repository
	kbRoleARN               string
	rdsCredentialsSecretARN string
	rdsAuroraClusterARN     string
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(
	cfg aws.Config,
	repository repository.Repository,
	kbRoleARN,
	rdsCredentialsSecretARN,
	rdsAuroraClusterARN string,
) RAGBuilder {
	return &BedrockRAGBuilder{
		s3Client:                s3.NewFromConfig(cfg),
		kbClient:                bedrockagent.NewFromConfig(cfg),
		rdsClient:               rdsdata.NewFromConfig(cfg),
		repository:              repository,
		kbRoleARN:               kbRoleARN,
		rdsCredentialsSecretARN: rdsCredentialsSecretARN,
		rdsAuroraClusterARN:     rdsAuroraClusterARN,
	}
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(ctx context.Context, _ repository.Repository) (RAGMetadata, error) {
	// Create RDS Aurora table if it doesn't exist
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id VARCHAR(255) PRIMARY KEY,
            text TEXT,
            embedding VECTOR,
            metadata JSON
        )
    `, b.GetRDSVectorTableName())

	_, err := b.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(b.rdsAuroraClusterARN),
		SecretArn:   aws.String(b.rdsCredentialsSecretARN),
		Database:    aws.String(RDSAuroraDatabaseName),
		Sql:         aws.String(createTableSQL),
	})
	if err != nil {
		return RAGMetadata{}, fmt.Errorf("failed to create/check RDS Aurora table: %w", err)
	}

	// Create Bedrock Knowledge Base
	kbOutput, err := b.kbClient.CreateKnowledgeBase(ctx, &bedrockagent.CreateKnowledgeBaseInput{
		KnowledgeBaseConfiguration: &types.KnowledgeBaseConfiguration{
			Type: types.KnowledgeBaseTypeVector,
		},
		Name:        aws.String(CodeRefactoringKBName),
		RoleArn:     aws.String(b.kbRoleARN),
		Description: aws.String(CodeRefactoringKBDescription),
		StorageConfiguration: &types.StorageConfiguration{
			RdsConfiguration: &types.RdsConfiguration{
				CredentialsSecretArn: aws.String(b.rdsCredentialsSecretARN),
				DatabaseName:         aws.String(RDSAuroraDatabaseName),
				FieldMapping: &types.RdsFieldMapping{
					PrimaryKeyField: aws.String("id"),
					TextField:       aws.String("text"),
					VectorField:     aws.String("embedding"),
					MetadataField:   aws.String("metadata"),
				},
				ResourceArn: aws.String(b.rdsAuroraClusterARN),
				TableName:   aws.String(b.GetRDSVectorTableName()),
			},
		},
	})
	if err != nil {
		return RAGMetadata{}, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return RAGMetadata{
		VectorStoreID: *kbOutput.KnowledgeBase.KnowledgeBaseArn,
		DataLocation:  b.repository.GetPath(),
		Provider:      ProviderName,
	}, nil
}

// TearDown implements RAGBuilder.
func (b BedrockRAGBuilder) TearDown(ctx context.Context, vectorStoreID string) error {
	// Delete the Bedrock Knowledge Base
	_, err := b.kbClient.DeleteKnowledgeBase(ctx, &bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(vectorStoreID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	// Optionally, you can also drop the RDS table if needed
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", b.GetRDSVectorTableName())
	_, err = b.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(b.rdsAuroraClusterARN),
		SecretArn:   aws.String(b.rdsCredentialsSecretARN),
		Database:    aws.String(RDSAuroraDatabaseName),
		Sql:         aws.String(dropTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to drop RDS Aurora table: %w", err)
	}

	return nil
}

// GetRDSVectorTableName returns the name of the RDS table used for vector storage.
func (b BedrockRAGBuilder) GetRDSVectorTableName() string {
	return b.repository.GetPath()
}
