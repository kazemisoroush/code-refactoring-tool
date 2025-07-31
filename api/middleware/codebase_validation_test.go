// Package middleware provides validation tests for the codebase entity
package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

func TestCodebaseValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CreateCodebaseRequest_ValidRequest", func(t *testing.T) {
		// Valid request should pass validation
		request := models.CreateCodebaseRequest{
			ProjectID:     "proj-12345",
			Name:          "My Codebase",
			Provider:      models.ProviderGitHub,
			URL:           "https://github.com/user/repo.git",
			DefaultBranch: "main",
			Tags: map[string]string{
				"env":  "production",
				"team": "backend",
			},
		}

		router := gin.New()
		router.POST("/projects/:project_id/codebases", NewCombinedValidationMiddleware[models.CreateCodebaseRequest]().Handle(), func(c *gin.Context) {
			validatedRequest, exists := GetValidatedRequest[models.CreateCodebaseRequest](c)
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
				return
			}
			c.JSON(http.StatusOK, validatedRequest)
		})

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/projects/proj-12345/codebases", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CreateCodebaseRequest_InvalidProvider", func(t *testing.T) {
		// Invalid provider should fail validation
		request := map[string]interface{}{
			"projectId":     "proj-12345",
			"name":          "My Codebase",
			"provider":      "invalid-provider",
			"url":           "https://github.com/user/repo.git",
			"defaultBranch": "main",
		}

		router := gin.New()
		router.POST("/projects/:project_id/codebases", NewCombinedValidationMiddleware[models.CreateCodebaseRequest]().Handle(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/projects/proj-12345/codebases", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "Provider must be one of: github, gitlab, bitbucket, custom")
	})

	t.Run("CreateCodebaseRequest_InvalidURL", func(t *testing.T) {
		// Invalid URL should fail validation
		request := models.CreateCodebaseRequest{
			ProjectID:     "proj-12345",
			Name:          "My Codebase",
			Provider:      models.ProviderGitHub,
			URL:           "not-a-valid-url",
			DefaultBranch: "main",
		}

		router := gin.New()
		router.POST("/projects/:project_id/codebases", NewCombinedValidationMiddleware[models.CreateCodebaseRequest]().Handle(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/projects/proj-12345/codebases", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "Url is invalid")
	})

	t.Run("GetCodebaseRequest_ValidUUID", func(t *testing.T) {
		// Valid UUID should pass validation
		router := gin.New()
		router.GET("/codebases/:id", NewURIValidationMiddleware[models.GetCodebaseRequest]().Handle(), func(c *gin.Context) {
			validatedRequest, exists := GetValidatedRequest[models.GetCodebaseRequest](c)
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
				return
			}
			c.JSON(http.StatusOK, validatedRequest)
		})

		req := httptest.NewRequest("GET", "/codebases/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetCodebaseRequest_InvalidUUID", func(t *testing.T) {
		// Invalid UUID should fail validation
		router := gin.New()
		router.GET("/codebases/:id", NewURIValidationMiddleware[models.GetCodebaseRequest]().Handle(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		req := httptest.NewRequest("GET", "/codebases/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "CodebaseId is invalid")
	})

	t.Run("ListCodebasesRequest_ValidQuery", func(t *testing.T) {
		// Valid query parameters should pass validation
		router := gin.New()
		router.GET("/codebases", NewQueryValidationMiddleware[models.ListCodebasesRequest]().Handle(), func(c *gin.Context) {
			validatedRequest, exists := GetValidatedRequest[models.ListCodebasesRequest](c)
			if !exists {
				c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
				return
			}
			c.JSON(http.StatusOK, validatedRequest)
		})

		req := httptest.NewRequest("GET", "/codebases?project_id=proj-12345&max_results=10&tag_filter=env:production", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("ListCodebasesRequest_InvalidMaxResults", func(t *testing.T) {
		// Invalid max_results should fail validation
		router := gin.New()
		router.GET("/codebases", NewQueryValidationMiddleware[models.ListCodebasesRequest]().Handle(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})

		req := httptest.NewRequest("GET", "/codebases?max_results=150", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"].(string), "Validation failed")
		assert.Contains(t, response["details"].(string), "MaxResults must be at most 100")
	})
}
