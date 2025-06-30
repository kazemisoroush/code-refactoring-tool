// Package storage provides functions to interact with AWS S3 for uploading and deleting files.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	bedrocktypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	// DataSourceName is the name of the data source used for the code refactoring tool.
	DataSourceName = "code-refactoring-tool-data-source"

	// DataSourceDescription is the description of the data source used for the code refactoring tool.
	DataSourceDescription = "Data source for the code refactoring tool knowledge base. This data source is used to store the codebase and other relevant files for the RAG pipeline."
)

// S3Storage implements the Storage interface for AWS S3.
type S3Storage struct {
	s3Client   *s3.Client
	bucketName string
	client     *bedrockagent.Client
}

// NewS3Storage creates a new S3Storage instance with the provided bucket name.
func NewS3Storage(awsConfig aws.Config, bucketName string) DataStore {
	return &S3Storage{
		s3Client:   s3.NewFromConfig(awsConfig),
		bucketName: bucketName,
		client:     bedrockagent.NewFromConfig(awsConfig),
	}
}

// Create checks if the S3 bucket exists and is accessible.
func (s S3Storage) Create(ctx context.Context, ragID string) error {
	// S3 does not require explicit creation of a bucket, but we can check if it exists.
	_, err := s.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to access bucket %s: %w", s.bucketName, err)
	}

	bucketARN := fmt.Sprintf("arn:aws:s3:::%s", s.bucketName)

	// Create bedrock data store
	_, err = s.client.CreateDataSource(ctx, &bedrockagent.CreateDataSourceInput{
		Name:               aws.String(DataSourceName),
		KnowledgeBaseId:    aws.String(ragID),
		DataDeletionPolicy: bedrocktypes.DataDeletionPolicyDelete,
		DataSourceConfiguration: &bedrocktypes.DataSourceConfiguration{
			S3Configuration: &bedrocktypes.S3DataSourceConfiguration{
				BucketArn: aws.String(bucketARN),
			},
		},
		VectorIngestionConfiguration: &bedrocktypes.VectorIngestionConfiguration{
			// TODO: Fix chunking here...
			ChunkingConfiguration: &bedrocktypes.ChunkingConfiguration{
				ChunkingStrategy: bedrocktypes.ChunkingStrategyFixedSize,
				FixedSizeChunkingConfiguration: &bedrocktypes.FixedSizeChunkingConfiguration{
					MaxTokens:         aws.Int32(300),
					OverlapPercentage: aws.Int32(50),
				},
			},
			// TODO: Look into these settings...
			// ContextEnrichmentConfiguration
			// CustomTransformationConfiguration
			// ParsingConfiguration
		},
		Description: aws.String(DataSourceDescription),
	})
	if err != nil {
		return fmt.Errorf("failed to create data source: %w", err)
	}

	return nil
}

// Detele deletes the data source from the knowledge base.
func (s S3Storage) Detele(ctx context.Context, dataSourceID string, ragID string) error {
	_, err := s.client.DeleteDataSource(ctx, &bedrockagent.DeleteDataSourceInput{
		DataSourceId:    aws.String(dataSourceID),
		KnowledgeBaseId: aws.String(ragID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete data source: %w", err)
	}

	return nil
}

// UploadDirectory uploads all files in a directory to S3 under the given prefix.
func (s S3Storage) UploadDirectory(ctx context.Context, localPath, remotePath string) error {
	uploader := manager.NewUploader(s.s3Client)
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}
		key := filepath.ToSlash(filepath.Join(remotePath, relPath))
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
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
			Body:   f,
		})
		return fmt.Errorf("failed to upload %s to S3: %w", key, err)
	})
}

// DeleteDirectory deletes all objects under a given prefix in the bucket.
func (s S3Storage) DeleteDirectory(ctx context.Context, prefix string) error {
	var toDelete []types.ObjectIdentifier
	paginator := s3.NewListObjectsV2Paginator(s.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects in bucket %s with prefix %s: %w", s.bucketName, prefix, err)
		}
		for _, obj := range page.Contents {
			toDelete = append(toDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}
	if len(toDelete) == 0 {
		return nil
	}
	_, err := s.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucketName),
		Delete: &types.Delete{Objects: toDelete},
	})
	return fmt.Errorf("failed to delete objects in bucket %s with prefix %s: %w", s.bucketName, prefix, err)
}
