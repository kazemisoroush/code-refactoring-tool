// Package vector provides an implementation of a vector data store using AWS RDS Aurora.
package vector

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
)

const (
	// RDSAuroraDatabaseName is the name of the RDS Aurora database.
	RDSAuroraDatabaseName = "RefactorVectorDb"
)

// RDSVectorStorage is an interface for vector data stores.
type RDSVectorStorage struct {
	rdsClient               *rdsdata.Client
	rdsAuroraClusterARN     string
	rdsCredentialsSecretARN string
}

// NewRDSVectorStore creates a new instance of RDSVectorStore with the provided AWS configuration and parameters.
func NewRDSVectorStore(
	awsConfig aws.Config,
	rdsAuroraClusterARN string,
	rdsCredentialsSecretARN string,
) Storage {
	return &RDSVectorStorage{
		rdsClient:               rdsdata.NewFromConfig(awsConfig),
		rdsAuroraClusterARN:     rdsAuroraClusterARN,
		rdsCredentialsSecretARN: rdsCredentialsSecretARN,
	}
}

// EnsureSchema implements VectorStore.
func (r *RDSVectorStorage) EnsureSchema(ctx context.Context, tableName string) error {
	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id VARCHAR(255) PRIMARY KEY,
            text TEXT,
            embedding VECTOR,
            metadata JSON
        )
    `, tableName)

	_, err := r.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(r.rdsAuroraClusterARN),
		SecretArn:   aws.String(r.rdsCredentialsSecretARN),
		Database:    aws.String(RDSAuroraDatabaseName),
		Sql:         aws.String(createTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to create/check RDS Aurora table: %w", err)
	}

	return nil
}

// DropSchema implements VectorStore.
func (r *RDSVectorStorage) DropSchema(ctx context.Context, tableName string) error {
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := r.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(r.rdsAuroraClusterARN),
		SecretArn:   aws.String(r.rdsCredentialsSecretARN),
		Database:    aws.String(RDSAuroraDatabaseName),
		Sql:         aws.String(dropTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to drop RDS Aurora table: %w", err)
	}

	return nil
}
