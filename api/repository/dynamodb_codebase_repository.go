// Package repository provides DynamoDB implementation for CodebaseRepository
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// DynamoDBCodebaseRepository implements CodebaseRepository for AWS DynamoDB
type DynamoDBCodebaseRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBCodebaseRepository creates a new DynamoDBCodebaseRepository
// NewDynamoDBCodebaseRepository creates a new DynamoDBCodebaseRepository
func NewDynamoDBCodebaseRepository(awsConfig aws.Config, tableName string) CodebaseRepository {
	if tableName == "" {
		tableName = "code-refactor-codebases"
	}
	client := dynamodb.NewFromConfig(awsConfig)
	return &DynamoDBCodebaseRepository{
		client:    client,
		tableName: tableName,
	}
}

// CreateCodebase creates a new codebase record in DynamoDB
func (r *DynamoDBCodebaseRepository) CreateCodebase(ctx context.Context, codebase *models.Codebase) error {
	item, err := attributevalue.MarshalMap(codebase)
	if err != nil {
		return fmt.Errorf("failed to marshal codebase record: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(codebase_id)"),
	})
	if err != nil {
		return fmt.Errorf("failed to create codebase in DynamoDB: %w", err)
	}
	return nil
}

// GetCodebase retrieves a codebase by ID from DynamoDB
func (r *DynamoDBCodebaseRepository) GetCodebase(ctx context.Context, codebaseID string) (*models.Codebase, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"codebase_id": &types.AttributeValueMemberS{Value: codebaseID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get codebase from DynamoDB: %w", err)
	}
	if result.Item == nil {
		return nil, nil // Not found
	}
	var codebase models.Codebase
	err = attributevalue.UnmarshalMap(result.Item, &codebase)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal codebase record: %w", err)
	}
	return &codebase, nil
}

// UpdateCodebase updates an existing codebase in DynamoDB
func (r *DynamoDBCodebaseRepository) UpdateCodebase(ctx context.Context, codebase *models.Codebase) error {
	item, err := attributevalue.MarshalMap(codebase)
	if err != nil {
		return fmt.Errorf("failed to marshal codebase record: %w", err)
	}
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(codebase_id)"),
	})
	if err != nil {
		return fmt.Errorf("failed to update codebase in DynamoDB: %w", err)
	}
	return nil
}

// DeleteCodebase deletes a codebase by ID from DynamoDB
func (r *DynamoDBCodebaseRepository) DeleteCodebase(ctx context.Context, codebaseID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"codebase_id": &types.AttributeValueMemberS{Value: codebaseID},
		},
		ConditionExpression: aws.String("attribute_exists(codebase_id)"),
	})
	if err != nil {
		return fmt.Errorf("failed to delete codebase from DynamoDB: %w", err)
	}
	return nil
}

// ListCodebases lists codebases with optional filtering and pagination from DynamoDB
func (r *DynamoDBCodebaseRepository) ListCodebases(ctx context.Context, filter CodebaseFilter) ([]*models.Codebase, string, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}
	if filter.MaxResults != nil {
		input.Limit = aws.Int32(int32(*filter.MaxResults))
	}
	if filter.NextToken != nil && *filter.NextToken != "" {
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"codebase_id": &types.AttributeValueMemberS{Value: *filter.NextToken},
		}
	}
	// Filtering by ProjectID or Provider
	var filterExprs []string
	exprAttrNames := map[string]string{}
	exprAttrValues := map[string]types.AttributeValue{}
	if filter.ProjectID != nil {
		filterExprs = append(filterExprs, "project_id = :project_id")
		exprAttrValues[":project_id"] = &types.AttributeValueMemberS{Value: *filter.ProjectID}
	}
	if filter.Provider != nil {
		filterExprs = append(filterExprs, "provider = :provider")
		exprAttrValues[":provider"] = &types.AttributeValueMemberS{Value: string(*filter.Provider)}
	}
	if filter.TagFilter != nil {
		filterExprs = append(filterExprs, "contains(tags, :tag)")
		exprAttrValues[":tag"] = &types.AttributeValueMemberS{Value: *filter.TagFilter}
	}
	if len(filterExprs) > 0 {
		input.FilterExpression = aws.String(strings.Join(filterExprs, " AND "))
		input.ExpressionAttributeValues = exprAttrValues
		if len(exprAttrNames) > 0 {
			input.ExpressionAttributeNames = exprAttrNames
		}
	}
	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list codebases: %w", err)
	}
	var codebases []*models.Codebase
	for _, item := range result.Items {
		var codebase models.Codebase
		err = attributevalue.UnmarshalMap(item, &codebase)
		if err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal codebase record: %w", err)
		}
		codebases = append(codebases, &codebase)
	}
	var nextToken string
	if result.LastEvaluatedKey != nil {
		if codebaseID, ok := result.LastEvaluatedKey["codebase_id"]; ok {
			if s, ok := codebaseID.(*types.AttributeValueMemberS); ok {
				nextToken = s.Value
			}
		}
	}
	return codebases, nextToken, nil
}

// CodebaseExists checks if a codebase exists in DynamoDB
func (r *DynamoDBCodebaseRepository) CodebaseExists(ctx context.Context, codebaseID string) (bool, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"codebase_id": &types.AttributeValueMemberS{Value: codebaseID},
		},
		ProjectionExpression: aws.String("codebase_id"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check codebase existence: %w", err)
	}
	return result.Item != nil, nil
}

// GetCodebasesByProject gets all codebases for a specific project from DynamoDB
func (r *DynamoDBCodebaseRepository) GetCodebasesByProject(ctx context.Context, projectID string) ([]*models.Codebase, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("project_id = :project_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":project_id": &types.AttributeValueMemberS{Value: projectID},
		},
	}
	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get codebases by project: %w", err)
	}
	var codebases []*models.Codebase
	for _, item := range result.Items {
		var codebase models.Codebase
		err = attributevalue.UnmarshalMap(item, &codebase)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal codebase record: %w", err)
		}
		codebases = append(codebases, &codebase)
	}
	return codebases, nil
}
