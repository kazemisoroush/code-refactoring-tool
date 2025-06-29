// Package storage provides an interface for uploading and deleting directories in a storage system.
package storage

import "context"

// Storage interface defines methods for uploading and deleting directories in a storage system.
//
//go:generate mockgen -destination=./mocks/mock_storage.go -mock_names=Storage=MockStorage -package=mocks . Storage
type Storage interface {
	// UploadDirectory uploads a local directory to a remote path in the storage system.
	UploadDirectory(ctx context.Context, localPath, remotePath string) error

	// DeleteDirectory deletes a remote directory in the storage system.
	DeleteDirectory(ctx context.Context, remotePath string) error
}
