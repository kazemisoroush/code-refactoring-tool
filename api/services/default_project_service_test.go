package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	repositoryMocks "github.com/kazemisoroush/code-refactoring-tool/api/repository/mocks"
)

func TestNewDefaultProjectService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.projectRepo)
}

func TestDefaultProjectService_CreateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	description := "Test project description"
	language := "go"
	request := models.CreateProjectRequest{
		Name:        "test-project",
		Description: &description,
		Language:    &language,
		Tags: map[string]string{
			"env":  "test",
			"team": "backend",
		},
	}

	// Mock the repository call
	mockRepo.EXPECT().
		CreateProject(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, record *repository.ProjectRecord) error {
			// Verify the project record fields
			assert.Equal(t, request.Name, record.Name)
			assert.Equal(t, request.Description, record.Description)
			assert.Equal(t, request.Language, record.Language)
			assert.Equal(t, request.Tags, record.Tags)
			assert.Equal(t, string(models.ProjectStatusActive), record.Status)
			assert.NotEmpty(t, record.ProjectID)
			assert.True(t, record.ProjectID[:5] == "proj-")
			return nil
		}).
		Times(1)

	response, err := service.CreateProject(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.NotEmpty(t, response.ProjectID)
	assert.True(t, response.ProjectID[:5] == "proj-")
	assert.NotEmpty(t, response.CreatedAt)
}

func TestDefaultProjectService_CreateProject_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	request := models.CreateProjectRequest{
		Name: "test-project",
	}

	repoError := errors.New("database error")
	mockRepo.EXPECT().
		CreateProject(gomock.Any(), gomock.Any()).
		Return(repoError).
		Times(1)

	response, err := service.CreateProject(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create project")
}

func TestDefaultProjectService_GetProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "proj-12345-abcde"
	description := "Test project"
	language := "go"
	now := time.Now().UTC()

	projectRecord := &repository.ProjectRecord{
		ProjectID:   projectID,
		Name:        "test-project",
		Description: &description,
		Language:    &language,
		Status:      string(models.ProjectStatusActive),
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags: map[string]string{
			"env": "test",
		},
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	mockRepo.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(projectRecord, nil).
		Times(1)

	response, err := service.GetProject(context.Background(), projectID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, projectID, response.ProjectID)
	assert.Equal(t, "test-project", response.Name)
	assert.Equal(t, &description, response.Description)
	assert.Equal(t, &language, response.Language)
}

func TestDefaultProjectService_GetProject_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "nonexistent-project"

	mockRepo.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(nil, nil).
		Times(1)

	response, err := service.GetProject(context.Background(), projectID)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "project not found")
}

func TestDefaultProjectService_UpdateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "proj-12345-abcde"
	originalName := "original-project"
	updatedName := "updated-project"
	description := "Updated description"
	now := time.Now().UTC()

	// Existing project record
	existingRecord := &repository.ProjectRecord{
		ProjectID:   projectID,
		Name:        originalName,
		Description: nil,
		Status:      string(models.ProjectStatusActive),
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        make(map[string]string),
		Metadata:    make(map[string]string),
	}

	request := models.UpdateProjectRequest{
		ProjectID:   projectID,
		Name:        &updatedName,
		Description: &description,
		Tags: map[string]string{
			"env": "staging",
		},
	}

	// Mock getting the existing project
	mockRepo.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(existingRecord, nil).
		Times(1)

	// Mock updating the project
	mockRepo.EXPECT().
		UpdateProject(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, record *repository.ProjectRecord) error {
			// Verify the updated fields
			assert.Equal(t, projectID, record.ProjectID)
			assert.Equal(t, updatedName, record.Name)
			assert.Equal(t, &description, record.Description)
			assert.Equal(t, request.Tags, record.Tags)
			assert.True(t, record.UpdatedAt.After(now))
			return nil
		}).
		Times(1)

	response, err := service.UpdateProject(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, projectID, response.ProjectID)
	assert.NotEmpty(t, response.UpdatedAt)
}

func TestDefaultProjectService_UpdateProject_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "nonexistent-project"
	updatedName := "updated-project"

	request := models.UpdateProjectRequest{
		ProjectID: projectID,
		Name:      &updatedName,
	}

	mockRepo.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(nil, nil).
		Times(1)

	response, err := service.UpdateProject(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "project not found")
}

func TestDefaultProjectService_DeleteProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "proj-12345-abcde"

	// Mock checking if project exists
	mockRepo.EXPECT().
		ProjectExists(gomock.Any(), projectID).
		Return(true, nil).
		Times(1)

	// Mock deleting the project
	mockRepo.EXPECT().
		DeleteProject(gomock.Any(), projectID).
		Return(nil).
		Times(1)

	response, err := service.DeleteProject(context.Background(), projectID)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Success)
}

func TestDefaultProjectService_DeleteProject_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	projectID := "nonexistent-project"

	mockRepo.EXPECT().
		ProjectExists(gomock.Any(), projectID).
		Return(false, nil).
		Times(1)

	response, err := service.DeleteProject(context.Background(), projectID)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "project not found")
}

func TestDefaultProjectService_ListProjects_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	maxResults := 10
	nextToken := "next-token"
	request := models.ListProjectsRequest{
		MaxResults: &maxResults,
		NextToken:  &nextToken,
		TagFilter: map[string]string{
			"env": "test",
		},
	}

	now := time.Now().UTC()
	projectRecords := []*repository.ProjectRecord{
		{
			ProjectID: "proj-12345-abcde",
			Name:      "project-1",
			CreatedAt: now,
			Tags: map[string]string{
				"env": "test",
			},
		},
		{
			ProjectID: "proj-67890-fghij",
			Name:      "project-2",
			CreatedAt: now.Add(-time.Hour),
			Tags: map[string]string{
				"env": "test",
			},
		},
	}

	returnNextToken := "return-next-token"

	mockRepo.EXPECT().
		ListProjects(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, opts repository.ListProjectsOptions) ([]*repository.ProjectRecord, string, error) {
			// Verify the options
			assert.Equal(t, request.NextToken, opts.NextToken)
			assert.Equal(t, request.MaxResults, opts.MaxResults)
			assert.Equal(t, request.TagFilter, opts.TagFilter)
			return projectRecords, returnNextToken, nil
		}).
		Times(1)

	response, err := service.ListProjects(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Projects, 2)
	assert.Equal(t, "proj-12345-abcde", response.Projects[0].ProjectID)
	assert.Equal(t, "project-1", response.Projects[0].Name)
	assert.Equal(t, "proj-67890-fghij", response.Projects[1].ProjectID)
	assert.Equal(t, "project-2", response.Projects[1].Name)
	assert.NotNil(t, response.NextToken)
	assert.Equal(t, returnNextToken, *response.NextToken)
}

func TestDefaultProjectService_ListProjects_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repositoryMocks.NewMockProjectRepository(ctrl)
	service := NewDefaultProjectService(mockRepo)

	request := models.ListProjectsRequest{}

	mockRepo.EXPECT().
		ListProjects(gomock.Any(), gomock.Any()).
		Return([]*repository.ProjectRecord{}, "", nil).
		Times(1)

	response, err := service.ListProjects(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Projects, 0)
	assert.Nil(t, response.NextToken)
}
