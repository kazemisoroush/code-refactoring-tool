// Package storage provides an implementation of a vector data store using AWS RDS Aurora.
package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// RDSVector is an interface for vector data stores.
type RDSVector struct {
	rdsClient               *rdsdata.Client
	rdsAuroraClusterARN     string
	rdsCredentialsSecretARN string
	rdsAuroraDatabaseName   string
}

// NewRDSVector creates a new instance of RDSVectorStore with the provided AWS configuration and parameters.
func NewRDSVector(
	awsConfig aws.Config,
	rdsAurora config.RDSAurora,
) Vector {
	return &RDSVector{
		rdsClient:               rdsdata.NewFromConfig(awsConfig),
		rdsAuroraClusterARN:     rdsAurora.ClusterARN,
		rdsCredentialsSecretARN: rdsAurora.CredentialsSecretARN,
		rdsAuroraDatabaseName:   rdsAurora.DatabaseName,
	}
}

// EnsureSchema implements VectorStore.
func (r *RDSVector) EnsureSchema(ctx context.Context, tableName string) error {
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
		Database:    aws.String(r.rdsAuroraDatabaseName),
		Sql:         aws.String(createTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to create/check RDS Aurora table: %w", err)
	}

	return nil
}

// DropSchema implements VectorStore.
func (r *RDSVector) DropSchema(ctx context.Context, tableName string) error {
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := r.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(r.rdsAuroraClusterARN),
		SecretArn:   aws.String(r.rdsCredentialsSecretARN),
		Database:    aws.String(r.rdsAuroraDatabaseName),
		Sql:         aws.String(dropTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to drop RDS Aurora table: %w", err)
	}

	return nil
}
