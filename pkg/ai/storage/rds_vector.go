// Package storage provides an implementation of a vector data store using AWS RDS Postgres.
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
	RDSPostgresInstanceARN  string
	rdsCredentialsSecretARN string
	rdsPostgresDatabaseName string
}

// NewRDSVector creates a new instance of RDSVectorStore with the provided AWS configuration and parameters.
func NewRDSVector(
	awsConfig aws.Config,
	rdsPostgres config.RDSPostgres,
) Vector {
	return &RDSVector{
		rdsClient:               rdsdata.NewFromConfig(awsConfig),
		RDSPostgresInstanceARN:  rdsPostgres.ClusterARN,
		rdsCredentialsSecretARN: rdsPostgres.CredentialsSecretARN,
		rdsPostgresDatabaseName: rdsPostgres.DatabaseName,
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
		ResourceArn: aws.String(r.RDSPostgresInstanceARN),
		SecretArn:   aws.String(r.rdsCredentialsSecretARN),
		Database:    aws.String(r.rdsPostgresDatabaseName),
		Sql:         aws.String(createTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to create/check RDS Postgres table: %w", err)
	}

	return nil
}

// DropSchema implements VectorStore.
func (r *RDSVector) DropSchema(ctx context.Context, tableName string) error {
	dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := r.rdsClient.ExecuteStatement(ctx, &rdsdata.ExecuteStatementInput{
		ResourceArn: aws.String(r.RDSPostgresInstanceARN),
		SecretArn:   aws.String(r.rdsCredentialsSecretARN),
		Database:    aws.String(r.rdsPostgresDatabaseName),
		Sql:         aws.String(dropTableSQL),
	})
	if err != nil {
		return fmt.Errorf("failed to drop RDS Postgres table: %w", err)
	}

	return nil
}
