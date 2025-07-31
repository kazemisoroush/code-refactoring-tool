package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	servicesMocks "github.com/kazemisoroush/code-refactoring-tool/api/services/mocks"
)

func TestNewProjectController(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	assert.NotNil(t, controller)
	assert.Equal(t, mockService, controller.projectService)
}

func TestProjectController_CreateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	description := "A sample project for testing"
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

	expectedResponse := &models.CreateProjectResponse{
		ProjectID: "proj-12345-abcde",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	mockService.EXPECT().
		CreateProject(gomock.Any(), request).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.POST("/projects", middleware.NewJSONValidationMiddleware[models.CreateProjectRequest]().Handle(), controller.CreateProject)

	// Create HTTP request
	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CreateProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ProjectID, response.ProjectID)
	assert.Equal(t, expectedResponse.CreatedAt, response.CreatedAt)
}

func TestProjectController_CreateProject_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	// Invalid JSON - missing required name field
	invalidJSON := `{"description": "missing name field"}`

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.POST("/projects", middleware.NewJSONValidationMiddleware[models.CreateProjectRequest]().Handle(), controller.CreateProject)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Equal(t, "Validation failed", errorResponse.Message)
	assert.Contains(t, errorResponse.Details, "Name is required")
}

func TestProjectController_CreateProject_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	request := models.CreateProjectRequest{
		Name: "test-project",
	}

	serviceError := errors.New("service error")
	mockService.EXPECT().
		CreateProject(gomock.Any(), request).
		Return(nil, serviceError).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.POST("/projects", middleware.NewJSONValidationMiddleware[models.CreateProjectRequest]().Handle(), controller.CreateProject)

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	assert.Equal(t, "Failed to create project", errorResponse.Message)
}

func TestProjectController_GetProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	projectID := "proj-12345-abcde"
	description := "A sample project"
	language := "go"
	expectedResponse := &models.GetProjectResponse{
		ProjectID:   projectID,
		Name:        "test-project",
		Description: &description,
		Language:    &language,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		Tags: map[string]string{
			"env": "test",
		},
		Metadata: map[string]string{
			"version": "1.0.0",
		},
	}

	mockService.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.GET("/projects/:id", middleware.NewURIValidationMiddleware[models.GetProjectRequest]().Handle(), controller.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ProjectID, response.ProjectID)
	assert.Equal(t, expectedResponse.Name, response.Name)
}

func TestProjectController_GetProject_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	// Use an invalid project ID that will fail validation
	projectID := "invalid-id"

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.GET("/projects/:id", middleware.NewURIValidationMiddleware[models.GetProjectRequest]().Handle(), controller.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Equal(t, "Validation failed", errorResponse.Message)
	assert.Contains(t, errorResponse.Details, "ID must start with 'proj-'")
}

func TestProjectController_GetProject_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	projectID := "proj-12345-abcde"
	notFoundError := errors.New("project not found")

	mockService.EXPECT().
		GetProject(gomock.Any(), projectID).
		Return(nil, notFoundError).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.GET("/projects/:id", middleware.NewURIValidationMiddleware[models.GetProjectRequest]().Handle(), controller.GetProject)

	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, errorResponse.Code)
	assert.Equal(t, "Project not found", errorResponse.Message)
}

func TestProjectController_UpdateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	projectID := "proj-12345-abcde"
	updatedName := "updated-project"
	request := models.UpdateProjectRequest{
		ProjectID: projectID,
		Name:      &updatedName,
		Tags: map[string]string{
			"env": "staging",
		},
	}

	expectedResponse := &models.UpdateProjectResponse{
		ProjectID: projectID,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	mockService.EXPECT().
		UpdateProject(gomock.Any(), request).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.PUT("/projects/:id", middleware.NewCombinedValidationMiddleware[models.UpdateProjectRequest]().Handle(), controller.UpdateProject)

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/projects/"+projectID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.UpdateProjectResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ProjectID, response.ProjectID)
	assert.Equal(t, expectedResponse.UpdatedAt, response.UpdatedAt)
}

func TestProjectController_DeleteProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	projectID := "proj-12345-abcde"
	expectedResponse := &models.DeleteProjectResponse{
		Success: true,
	}

	mockService.EXPECT().
		DeleteProject(gomock.Any(), projectID).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.DELETE("/projects/:id", middleware.NewURIValidationMiddleware[models.DeleteProjectRequest]().Handle(), controller.DeleteProject)

	req := httptest.NewRequest(http.MethodDelete, "/projects/"+projectID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeleteProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.Success, response.Success)
}

func TestProjectController_ListProjects_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	request := models.ListProjectsRequest{
		MaxResults: func() *int { i := 10; return &i }(),
		// Note: TagFilter is complex to bind from query params, so we'll test without it for now
	}

	expectedProjects := []models.ProjectSummary{
		{
			ProjectID: "proj-12345-abcde",
			Name:      "test-project-1",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			Tags: map[string]string{
				"env": "test",
			},
		},
		{
			ProjectID: "proj-67890-fghij",
			Name:      "test-project-2",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			Tags: map[string]string{
				"env": "test",
			},
		},
	}

	nextToken := "next-token-123"
	expectedResponse := &models.ListProjectsResponse{
		Projects:  expectedProjects,
		NextToken: &nextToken,
	}

	mockService.EXPECT().
		ListProjects(gomock.Any(), request).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.GET("/projects", middleware.NewQueryValidationMiddleware[models.ListProjectsRequest]().Handle(), controller.ListProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects?max_results=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ListProjectsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Projects, 2)
	assert.Equal(t, expectedProjects[0].ProjectID, response.Projects[0].ProjectID)
	assert.Equal(t, expectedProjects[1].ProjectID, response.Projects[1].ProjectID)
	assert.NotNil(t, response.NextToken)
	assert.Equal(t, nextToken, *response.NextToken)
}

func TestProjectController_ListProjects_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockProjectService(ctrl)
	controller := NewProjectController(mockService)

	request := models.ListProjectsRequest{}

	expectedResponse := &models.ListProjectsResponse{
		Projects:  []models.ProjectSummary{},
		NextToken: nil,
	}

	mockService.EXPECT().
		ListProjects(gomock.Any(), request).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Use validation middleware with the controller
	router.GET("/projects", middleware.NewQueryValidationMiddleware[models.ListProjectsRequest]().Handle(), controller.ListProjects)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ListProjectsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Projects, 0)
	assert.Nil(t, response.NextToken)
}
