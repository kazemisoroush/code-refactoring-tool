// Package storage provides functions to interact with AWS S3 for uploading and deleting files.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements the Storage interface for AWS S3.
type S3Storage struct {
	s3Client   *s3.Client
	bucketName string
}

// NewS3Storage creates a new S3Storage instance with the provided bucket name.
func NewS3Storage(bucketName string) Storage {
	return &S3Storage{
		s3Client:   s3.NewFromConfig(aws.Config{}),
		bucketName: bucketName,
	}
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
