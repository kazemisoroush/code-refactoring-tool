package repository

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/stretchr/testify/assert"
)

func TestDynamoDBCodebaseRepository_NotImplemented(t *testing.T) {
	// Use zero-value aws.Config for testing
	repo := NewDynamoDBCodebaseRepository(aws.Config{}, "dummy-table")
	ctx := context.Background()

	_, err := repo.GetCodebase(ctx, "id")
	assert.Error(t, err)

	err = repo.CreateCodebase(ctx, &models.Codebase{})
	assert.Error(t, err)

	err = repo.UpdateCodebase(ctx, &models.Codebase{})
	assert.Error(t, err)

	err = repo.DeleteCodebase(ctx, "id")
	assert.Error(t, err)

	_, _, err = repo.ListCodebases(ctx, CodebaseFilter{})
	assert.Error(t, err)

	exists, err := repo.CodebaseExists(ctx, "id")
	assert.Error(t, err)
	assert.False(t, exists)

	_, err = repo.GetCodebasesByProject(ctx, "project-id")
	assert.Error(t, err)
}
