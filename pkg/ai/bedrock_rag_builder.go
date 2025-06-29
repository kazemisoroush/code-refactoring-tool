// Package ai contains interfaces and types for building AI agents based on RAG (Retrieval-Augmented Generation) metadata.
package ai

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	s3BucketName            string
	kbRoleARN               string
	rdsCredentialsSecretARN string
	rdsAuroraClusterARN     string
}

// NewBedrockRAGBuilder creates a new instance of BedrockRAGBuilder.
func NewBedrockRAGBuilder(
	cfg aws.Config,
	repository repository.Repository,
	s3BucketName,
	kbRoleARN,
	rdsCredentialsSecretARN,
	rdsAuroraClusterARN string,
) RAGBuilder {
	return &BedrockRAGBuilder{
		s3Client:                s3.NewFromConfig(cfg),
		kbClient:                bedrockagent.NewFromConfig(cfg),
		rdsClient:               rdsdata.NewFromConfig(cfg),
		repository:              repository,
		s3BucketName:            s3BucketName,
		kbRoleARN:               kbRoleARN,
		rdsCredentialsSecretARN: rdsCredentialsSecretARN,
		rdsAuroraClusterARN:     rdsAuroraClusterARN,
	}
}

// Build implements RAGBuilder.
func (b BedrockRAGBuilder) Build(ctx context.Context, _ repository.Repository) (RAGMetadata, error) {
	// TODO: Upload the codebase to S3 if not already done
	repoPath := b.repository.GetPath()
	err := uploadDirectoryToS3(ctx, b.s3Client, repoPath, b.s3BucketName, repoPath)
	if err != nil {
		return RAGMetadata{}, fmt.Errorf("failed to upload codebase to S3: %w", err)
	}

	// Create RDS Aurora table if it doesn't exist
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id VARCHAR(255) PRIMARY KEY,
            text TEXT,
            embedding VECTOR,
            metadata JSON
        )
    `, b.getRDSAuroraTableName())

	_, err = b.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
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
				TableName:   aws.String(b.getRDSAuroraTableName()),
			},
		},
	})
	if err != nil {
		return RAGMetadata{}, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return RAGMetadata{
		VectorStoreID: *kbOutput.KnowledgeBase.KnowledgeBaseArn,
		Provider:      ProviderName,
	}, nil
}

// TearDown implements RAGBuilder.
func (b BedrockRAGBuilder) TearDown(ctx context.Context, vectorStoreID string) error {
	// TODO: Remove the codebase from S3 if needed
	repoPath := b.repository.GetPath()
	err := deleteS3Prefix(ctx, b.s3Client, b.s3BucketName, repoPath)
	if err != nil {
		return fmt.Errorf("failed to remove codebase from S3: %w", err)
	}

	// Delete the Bedrock Knowledge Base
	_, err = b.kbClient.DeleteKnowledgeBase(ctx, &bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(vectorStoreID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	// Optionally, you can also drop the RDS table if needed
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", b.getRDSAuroraTableName())
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

// getRDSAuroraTableName returns the name of the RDS table used for vector storage.
func (b BedrockRAGBuilder) getRDSAuroraTableName() string {
	return b.repository.GetPath()
}

// uploadDirectoryToS3 uploads all files in a directory to S3 under the given prefix.
func uploadDirectoryToS3(ctx context.Context, s3Client *s3.Client, localDir, bucket, prefix string) error {
	uploader := manager.NewUploader(s3Client)
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		key := filepath.ToSlash(filepath.Join(prefix, relPath))
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "failed to close file %s: %v\n", path, cerr)
			}
		}()
		_, err = uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   f,
		})
		return fmt.Errorf("failed to upload %s to S3: %w", key, err)
	})
}

// deleteS3Prefix deletes all objects under a given prefix in the bucket.
func deleteS3Prefix(ctx context.Context, s3Client *s3.Client, bucket, prefix string) error {
	var toDelete []s3types.ObjectIdentifier
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket %s with prefix %s: %w", bucket, prefix, err)
		}
		for _, obj := range page.Contents {
			toDelete = append(toDelete, s3types.ObjectIdentifier{Key: obj.Key})
		}
	}
	if len(toDelete) == 0 {
		return nil
	}
	_, err := s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3types.Delete{Objects: toDelete},
	})
	return fmt.Errorf("failed to delete objects in bucket %s with prefix %s: %w", bucket, prefix, err)
}
