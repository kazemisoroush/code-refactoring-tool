package repository

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestNewDynamoDBAgentRepository(t *testing.T) {
	tests := []struct {
		name          string
		awsConfig     aws.Config
		tableName     string
		expectedTable string
	}{
		{
			name: "with custom table name",
			awsConfig: aws.Config{
				Region: "us-east-1",
			},
			tableName:     "custom-agents-table",
			expectedTable: "custom-agents-table",
		},
		{
			name: "with empty table name uses default",
			awsConfig: aws.Config{
				Region: "us-east-1",
			},
			tableName:     "",
			expectedTable: DefaultTableName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewDynamoDBAgentRepository(tt.awsConfig, tt.tableName)

			// Assert that the repository is created
			assert.NotNil(t, repo)

			// Cast to concrete type to check internal state
			dynamoRepo, ok := repo.(*DynamoDBAgentRepository)
			assert.True(t, ok, "Repository should be of type *DynamoDBAgentRepository")
			assert.Equal(t, tt.expectedTable, dynamoRepo.tableName)
			assert.NotNil(t, dynamoRepo.client)
		})
	}
}

func TestDynamoDBAgentRepository_DefaultTableName(t *testing.T) {
	awsConfig := aws.Config{
		Region: "us-west-2",
	}

	repo := NewDynamoDBAgentRepository(awsConfig, "") // Empty table name

	dynamoRepo := repo.(*DynamoDBAgentRepository)
	assert.Equal(t, DefaultTableName, dynamoRepo.tableName, "Should use default table name when none provided")
}
