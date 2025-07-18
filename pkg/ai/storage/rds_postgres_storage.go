// Package storage provides interfaces and clients for interacting with storage backends,
// including services like RDS for persistent schema storage and S3 for object storage.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// RDSPostgresStorage is a client that ensures a Postgres schema exists
// by invoking a Lambda function which connects to the RDS instance and
// creates the schema if it does not already exist.
type RDSPostgresStorage struct {
	client    *lambda.Client
	lambdaARN string
}

var response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NewRDSPostgresStorage initializes an RDSPostgresStorage using the given AWS config and Lambda ARN.
// The Lambda must accept a JSON payload of the form { "table": "<table_name>" }
// and ensure the schema is created in a Postgres database.
func NewRDSPostgresStorage(awsConfig aws.Config, lambdaARN string) Storage {
	return &RDSPostgresStorage{
		client:    lambda.NewFromConfig(awsConfig),
		lambdaARN: lambdaARN,
	}
}

// EnsureSchema triggers the Lambda function to create the schema/table in the RDS Postgres database.
// It sends the database name and table name in the request payload and parses the response to confirm success or capture errors.
func (c *RDSPostgresStorage) EnsureSchema(ctx context.Context, databaseName string, tableName string) error {
	payload := map[string]string{
		"database": databaseName,
		"table":    tableName,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling Lambda payload: %w", err)
	}

	output, err := c.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(c.lambdaARN),
		InvocationType: types.InvocationTypeRequestResponse,
		Payload:        data,
	})
	if err != nil {
		return fmt.Errorf("error invoking Lambda function: %w", err)
	}

	err = json.Unmarshal(output.Payload, &response)
	if err != nil {
		return fmt.Errorf("failed to parse Lambda response payload: %w", err)
	}

	if response.Status != "success" {
		return errors.New("Lambda error: " + response.Message)
	}

	fmt.Println("EnsureSchema success:", response.Message)
	return nil
}
