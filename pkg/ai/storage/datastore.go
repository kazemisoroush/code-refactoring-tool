// Package storage provides an interface for uploading and deleting directories in a storage system.
package storage

import "context"

// DataStore interface defines methods for uploading and deleting directories in a storage system.
//
//go:generate mockgen -destination=./mocks/mock_datastore.go -mock_names=DataStore=MockDataStore -package=mocks . DataStore
type DataStore interface {
	// Create initializes the data store, preparing it for use.
	Create(ctx context.Context, ragID string) (string, error)

	// Delete removes the data store, cleaning up any resources it holds.
	Delete(ctx context.Context, dataSourceID string, ragID string) error

	// UploadDirectory uploads a local directory to a remote path in the storage system.
	UploadDirectory(ctx context.Context, localPath, remotePath string) error

	// DeleteDirectory deletes a remote directory in the storage system.
	DeleteDirectory(ctx context.Context, remotePath string) error
}
