// Package routes provides the HTTP routes for the API.
package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services/mocks"
)

func TestProjectRoutesWithValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock service with proper setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockProjectService(ctrl)
	controller := controllers.NewProjectController(mockService)

	// Setup router with validation middleware
	router := gin.New()
	SetupProjectRoutes(router, controller)

	t.Run("CreateProject_ValidRequest", func(t *testing.T) {
		// Set up mock expectation for successful validation and creation
		mockService.EXPECT().
			CreateProject(gomock.Any(), gomock.Any()).
			Return(&models.CreateProjectResponse{
				ProjectID: "proj-12345",
				CreatedAt: "2024-01-01T00:00:00Z",
			}, nil)

		// Valid request should pass validation and reach controller
		request := models.CreateProjectRequest{
			Name:        "Test Project",
			Description: stringPtr("A test project"),
			Language:    stringPtr("go"),
			Tags:        map[string]string{"env": "test"},
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should reach the controller and return success
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateProject_InvalidRequest", func(t *testing.T) {
		// Invalid request should fail validation before reaching controller
		request := models.CreateProjectRequest{
			// Missing required name field
			Description: stringPtr("A test project"),
			Language:    stringPtr("go"),
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/projects", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return validation error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "Name is required")
	})

	t.Run("UpdateProject_ValidRequest", func(t *testing.T) {
		// Set up mock expectation for successful validation and update
		mockService.EXPECT().
			UpdateProject(gomock.Any(), gomock.Any()).
			Return(&models.UpdateProjectResponse{
				ProjectID: "proj-12345",
				UpdatedAt: "2024-01-01T00:00:00Z",
			}, nil)

		// Valid combined URI + JSON request should pass validation
		request := models.UpdateProjectRequest{
			Name:        stringPtr("Updated Project"),
			Description: stringPtr("Updated description"),
			Language:    stringPtr("python"),
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("PUT", "/api/v1/projects/proj-12345", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not return validation error
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateProject_InvalidProjectID", func(t *testing.T) {
		// Project ID that doesn't exist should fail at service layer
		request := models.UpdateProjectRequest{
			Name: stringPtr("Updated Project"),
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		// Set up mock to return not found error
		mockService.EXPECT().
			UpdateProject(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("project not found"))

		req := httptest.NewRequest("PUT", "/api/v1/projects/nonexistent-id", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return not found error from service
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ListProjects_ValidQuery", func(t *testing.T) {
		// Set up mock expectation for list operation
		mockService.EXPECT().
			ListProjects(gomock.Any(), gomock.Any()).
			Return(&models.ListProjectsResponse{
				Projects: []models.ProjectSummary{},
			}, nil)

		// Valid query parameters should pass validation
		req := httptest.NewRequest("GET", "/api/v1/projects?max_results=50&next_token=abc123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should reach the controller and return success
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListProjects_InvalidQuery", func(t *testing.T) {
		// Invalid query parameters should fail validation
		req := httptest.NewRequest("GET", "/api/v1/projects?max_results=150", nil) // max_results > 100
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return validation error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "MaxResults must be at most 100")
	})

	t.Run("GetProject_InvalidProjectID", func(t *testing.T) {
		// Set up mock to return not found error
		mockService.EXPECT().
			GetProject(gomock.Any(), "invalid-id").
			Return(nil, errors.New("project not found"))

		req := httptest.NewRequest("GET", "/api/v1/projects/invalid-id", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return not found error from service
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestValidationMiddlewareExtensibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CustomValidationRulesWork", func(t *testing.T) {
		// Test that our custom project_id validation rule works
		router := gin.New()

		router.GET("/test/:project_id", middleware.NewURIValidationMiddleware[models.GetProjectRequest]().Handle(), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Valid project ID - any non-empty string
		req := httptest.NewRequest("GET", "/test/any-project-id", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Another valid project ID
		req = httptest.NewRequest("GET", "/test/12345-abcdef", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GenericMiddlewareWorksWithDifferentTypes", func(t *testing.T) {
		// Define a test struct for another entity
		type TestEntityRequest struct {
			ID   string `uri:"id" validate:"required,project_id"`
			Name string `json:"name" validate:"required,min=1,max=50"`
		}

		router := gin.New()

		router.PUT("/entities/:id", middleware.NewCombinedValidationMiddleware[TestEntityRequest]().Handle(), func(c *gin.Context) {
			validated, exists := middleware.GetValidatedRequest[TestEntityRequest](c)
			assert.True(t, exists)
			c.JSON(200, gin.H{"id": validated.ID, "name": validated.Name})
		})

		// Valid request
		request := map[string]string{"name": "Test Entity"}
		jsonData, _ := json.Marshal(request)
		req := httptest.NewRequest("PUT", "/entities/proj-12345", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "proj-12345", response["id"])
		assert.Equal(t, "Test Entity", response["name"])
	})
}

func stringPtr(s string) *string {
	return &s
}
