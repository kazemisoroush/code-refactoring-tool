// Package repository provides DynamoDB implementation for project data operations
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	// DefaultProjectTableName is the default DynamoDB table name for projects
	DefaultProjectTableName = "code-refactor-projects"
)

// DynamoDBProjectRepository implements ProjectRepository using DynamoDB
type DynamoDBProjectRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBProjectRepository creates a new DynamoDB project repository
func NewDynamoDBProjectRepository(awsConfig aws.Config, tableName string) ProjectRepository {
	if tableName == "" {
		tableName = DefaultProjectTableName
	}

	return &DynamoDBProjectRepository{
		client:    dynamodb.NewFromConfig(awsConfig),
		tableName: tableName,
	}
}

// CreateProject creates a new project record in DynamoDB
func (r *DynamoDBProjectRepository) CreateProject(ctx context.Context, project *ProjectRecord) error {
	// Convert the project record to DynamoDB item
	item, err := attributevalue.MarshalMap(project)
	if err != nil {
		return fmt.Errorf("failed to marshal project record: %w", err)
	}

	// Add condition to ensure project doesn't already exist
	input := &dynamodb.PutItemInput{
		TableName:                   &r.tableName,
		Item:                        item,
		ConditionExpression:         aws.String("attribute_not_exists(project_id)"),
		ReturnConsumedCapacity:      types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsSize,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("project with ID %s already exists", project.ProjectID)
		}
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetProject retrieves a project by ID from DynamoDB
func (r *DynamoDBProjectRepository) GetProject(ctx context.Context, projectID string) (*ProjectRecord, error) {
	input := &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"project_id": &types.AttributeValueMemberS{Value: projectID},
		},
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Project not found
	}

	var project ProjectRecord
	err = attributevalue.UnmarshalMap(result.Item, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal project record: %w", err)
	}

	return &project, nil
}

// UpdateProject updates an existing project record in DynamoDB
func (r *DynamoDBProjectRepository) UpdateProject(ctx context.Context, project *ProjectRecord) error {
	// Convert the project record to DynamoDB item
	item, err := attributevalue.MarshalMap(project)
	if err != nil {
		return fmt.Errorf("failed to marshal project record: %w", err)
	}

	// Add condition to ensure project exists
	input := &dynamodb.PutItemInput{
		TableName:                   &r.tableName,
		Item:                        item,
		ConditionExpression:         aws.String("attribute_exists(project_id)"),
		ReturnConsumedCapacity:      types.ReturnConsumedCapacityTotal,
		ReturnItemCollectionMetrics: types.ReturnItemCollectionMetricsSize,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("project with ID %s does not exist", project.ProjectID)
		}
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// DeleteProject deletes a project by ID from DynamoDB
func (r *DynamoDBProjectRepository) DeleteProject(ctx context.Context, projectID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"project_id": &types.AttributeValueMemberS{Value: projectID},
		},
		ConditionExpression: aws.String("attribute_exists(project_id)"),
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return fmt.Errorf("project with ID %s does not exist", projectID)
		}
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ProjectExists checks if a project exists by ID
func (r *DynamoDBProjectRepository) ProjectExists(ctx context.Context, projectID string) (bool, error) {
	input := &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"project_id": &types.AttributeValueMemberS{Value: projectID},
		},
		ProjectionExpression: aws.String("project_id"),
	}

	result, err := r.client.GetItem(ctx, input)
	if err != nil {
		return false, fmt.Errorf("failed to check project existence: %w", err)
	}

	return result.Item != nil, nil
}

// ListProjects retrieves projects with pagination and filtering
func (r *DynamoDBProjectRepository) ListProjects(ctx context.Context, opts ListProjectsOptions) ([]*ProjectRecord, string, error) {
	input := &dynamodb.ScanInput{
		TableName: &r.tableName,
	}

	// Set limit if provided
	if opts.MaxResults != nil {
		input.Limit = aws.Int32(int32(*opts.MaxResults))
	}

	// Set starting point for pagination if provided
	if opts.NextToken != nil && *opts.NextToken != "" {
		// In a real implementation, you would decode the next token
		// For simplicity, we'll use it as the last evaluated key
		// This should be properly implemented with base64 encoding/decoding
		input.ExclusiveStartKey = map[string]types.AttributeValue{
			"project_id": &types.AttributeValueMemberS{Value: *opts.NextToken},
		}
	}

	// Add filter expression for tags if provided
	if len(opts.TagFilter) > 0 {
		filterExpression, expressionAttributeNames, expressionAttributeValues := buildTagFilterExpression(opts.TagFilter)
		input.FilterExpression = aws.String(filterExpression)
		input.ExpressionAttributeNames = expressionAttributeNames
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list projects: %w", err)
	}

	var projects []*ProjectRecord
	for _, item := range result.Items {
		var project ProjectRecord
		err = attributevalue.UnmarshalMap(item, &project)
		if err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal project record: %w", err)
		}
		projects = append(projects, &project)
	}

	var nextToken string
	if result.LastEvaluatedKey != nil {
		// In a real implementation, you would encode this properly
		if projectID, ok := result.LastEvaluatedKey["project_id"]; ok {
			if s, ok := projectID.(*types.AttributeValueMemberS); ok {
				nextToken = s.Value
			}
		}
	}

	return projects, nextToken, nil
}

// buildTagFilterExpression builds DynamoDB filter expression for tag filtering
func buildTagFilterExpression(tagFilter map[string]string) (string, map[string]string, map[string]types.AttributeValue) {
	if len(tagFilter) == 0 {
		return "", nil, nil
	}

	var filterConditions []string
	expressionAttributeNames := make(map[string]string)
	expressionAttributeValues := make(map[string]types.AttributeValue)

	i := 0
	for key, value := range tagFilter {
		keyPlaceholder := fmt.Sprintf("#tag_key_%d", i)
		valuePlaceholder := fmt.Sprintf(":tag_value_%d", i)

		// DynamoDB doesn't support nested map filtering easily
		// In a real implementation, you might store tags as separate items
		// or use a different approach. For now, we'll use contains check on JSON string
		condition := fmt.Sprintf("contains(tags, %s) AND contains(tags, %s)", keyPlaceholder, valuePlaceholder)
		filterConditions = append(filterConditions, condition)

		expressionAttributeNames[keyPlaceholder] = key
		expressionAttributeValues[valuePlaceholder] = &types.AttributeValueMemberS{Value: value}
		i++
	}

	filterExpression := "(" + filterConditions[0]
	for j := 1; j < len(filterConditions); j++ {
		filterExpression += " AND " + filterConditions[j]
	}
	filterExpression += ")"

	return filterExpression, expressionAttributeNames, expressionAttributeValues
}
