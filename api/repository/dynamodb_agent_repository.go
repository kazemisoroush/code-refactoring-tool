// Package repository provides DynamoDB implementation for agent data operations
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

const (
	// DefaultTableName is the default DynamoDB table name for agents
	DefaultTableName = "code-refactor-agents"
)

// DynamoDBAgentRepository implements AgentRepository using DynamoDB
type DynamoDBAgentRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBAgentRepository creates a new DynamoDB agent repository
func NewDynamoDBAgentRepository(awsConfig aws.Config, tableName string) AgentRepository {
	if tableName == "" {
		tableName = DefaultTableName
	}

	client := dynamodb.NewFromConfig(awsConfig)

	return &DynamoDBAgentRepository{
		client:    client,
		tableName: tableName,
	}
}

// CreateAgent stores a new agent record
func (r *DynamoDBAgentRepository) CreateAgent(ctx context.Context, agent *AgentRecord) error {
	item, err := attributevalue.MarshalMap(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent record: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(agent_id)"), // Prevent overwriting
	})
	if err != nil {
		return fmt.Errorf("failed to create agent in DynamoDB: %w", err)
	}

	return nil
}

// GetAgent retrieves an agent by ID
func (r *DynamoDBAgentRepository) GetAgent(ctx context.Context, agentID string) (*AgentRecord, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"agent_id": &types.AttributeValueMemberS{Value: agentID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get agent from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	var agent AgentRecord
	err = attributevalue.UnmarshalMap(result.Item, &agent)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent record: %w", err)
	}

	return &agent, nil
}

// UpdateAgent updates an existing agent record
func (r *DynamoDBAgentRepository) UpdateAgent(ctx context.Context, agent *AgentRecord) error {
	agent.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent record: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update agent in DynamoDB: %w", err)
	}

	return nil
}

// DeleteAgent removes an agent record
func (r *DynamoDBAgentRepository) DeleteAgent(ctx context.Context, agentID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"agent_id": &types.AttributeValueMemberS{Value: agentID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete agent from DynamoDB: %w", err)
	}

	return nil
}

// ListAgents retrieves all agent records
func (r *DynamoDBAgentRepository) ListAgents(ctx context.Context) ([]*AgentRecord, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan agents from DynamoDB: %w", err)
	}

	agents := make([]*AgentRecord, 0, len(result.Items))
	for _, item := range result.Items {
		var agent AgentRecord
		err = attributevalue.UnmarshalMap(item, &agent)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal agent record: %w", err)
		}
		agents = append(agents, &agent)
	}

	return agents, nil
}

// UpdateAgentStatus updates only the status field
func (r *DynamoDBAgentRepository) UpdateAgentStatus(ctx context.Context, agentID string, status models.AgentStatus) error {
	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"agent_id": &types.AttributeValueMemberS{Value: agentID},
		},
		UpdateExpression: aws.String("SET #status = :status, updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":     &types.AttributeValueMemberS{Value: string(status)},
			":updated_at": &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update agent status in DynamoDB: %w", err)
	}

	return nil
}
