// Package routes provides the HTTP routes for the API.
package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

func TestValidationIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("DirectValidationWorks", func(t *testing.T) {
		router := gin.New()

		router.POST("/test", middleware.ValidateJSON[models.CreateProjectRequest](), func(c *gin.Context) {
			validated, exists := middleware.GetValidatedRequest[models.CreateProjectRequest](c)
			require.True(t, exists)
			c.JSON(200, gin.H{"name": validated.Name})
		})

		// Valid request
		request := models.CreateProjectRequest{
			Name:        "Test Project",
			Description: stringPtr("A test project"),
			Language:    stringPtr("go"),
			Tags:        map[string]string{"env": "test"},
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Test Project", response["name"])
	})

	t.Run("ValidationCatchesErrors", func(t *testing.T) {
		router := gin.New()

		router.POST("/test", middleware.ValidateJSON[models.CreateProjectRequest](), func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "should not reach here"})
		})

		// Invalid request - missing name
		request := map[string]interface{}{
			"description": "A test project",
		}

		jsonData, err := json.Marshal(request)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["details"].(string), "Name is required")
	})
}
