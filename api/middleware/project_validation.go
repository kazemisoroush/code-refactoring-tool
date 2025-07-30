// Package middleware provides HTTP middleware for the API
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// ValidateCreateProject validates create project request
func ValidateCreateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.CreateProjectRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate required fields
		if strings.TrimSpace(request.Name) == "" {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Project name is required",
				Details: "Name field cannot be empty",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate name length
		if len(request.Name) > 100 {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Project name too long",
				Details: "Name must be 100 characters or less",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate description length if provided
		if request.Description != nil && len(*request.Description) > 500 {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Project description too long",
				Details: "Description must be 500 characters or less",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate language if provided
		if request.Language != nil {
			validLanguages := map[string]bool{
				"go":         true,
				"javascript": true,
				"typescript": true,
				"python":     true,
				"java":       true,
				"csharp":     true,
				"rust":       true,
				"cpp":        true,
				"c":          true,
				"ruby":       true,
				"php":        true,
				"kotlin":     true,
				"swift":      true,
				"scala":      true,
				"other":      true,
			}

			if !validLanguages[strings.ToLower(*request.Language)] {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid programming language",
					Details: "Language must be one of: go, javascript, typescript, python, java, csharp, rust, cpp, c, ruby, php, kotlin, swift, scala, other",
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Validate tags
		if err := validateTags(request.Tags); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid tags",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Store validated request in context for the handler
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// ValidateUpdateProject validates update project request
func ValidateUpdateProject() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.UpdateProjectRequest

		// Bind URI parameter
		if err := c.ShouldBindUri(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid project ID",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Bind JSON body
		if err := c.ShouldBindJSON(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate project ID format
		if !isValidProjectID(request.ProjectID) {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid project ID format",
				Details: "Project ID must start with 'proj-' followed by alphanumeric characters",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate name if provided
		if request.Name != nil {
			if strings.TrimSpace(*request.Name) == "" {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Project name cannot be empty",
					Details: "If provided, name field cannot be empty",
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}

			if len(*request.Name) > 100 {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Project name too long",
					Details: "Name must be 100 characters or less",
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Validate description if provided
		if request.Description != nil && len(*request.Description) > 500 {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Project description too long",
				Details: "Description must be 500 characters or less",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate language if provided
		if request.Language != nil {
			validLanguages := map[string]bool{
				"go":         true,
				"javascript": true,
				"typescript": true,
				"python":     true,
				"java":       true,
				"csharp":     true,
				"rust":       true,
				"cpp":        true,
				"c":          true,
				"ruby":       true,
				"php":        true,
				"kotlin":     true,
				"swift":      true,
				"scala":      true,
				"other":      true,
			}

			if !validLanguages[strings.ToLower(*request.Language)] {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid programming language",
					Details: "Language must be one of: go, javascript, typescript, python, java, csharp, rust, cpp, c, ruby, php, kotlin, swift, scala, other",
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Validate tags if provided
		if request.Tags != nil {
			if err := validateTags(request.Tags); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid tags",
					Details: err.Error(),
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Validate metadata if provided
		if request.Metadata != nil {
			if err := validateMetadata(request.Metadata); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid metadata",
					Details: err.Error(),
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Store validated request in context for the handler
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// ValidateProjectID validates project ID parameter
func ValidateProjectID() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		if !isValidProjectID(projectID) {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid project ID format",
				Details: "Project ID must start with 'proj-' followed by alphanumeric characters",
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateListProjectsQuery validates list projects query parameters
func ValidateListProjectsQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request models.ListProjectsRequest

		if err := c.ShouldBindQuery(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Validate max_results
		if request.MaxResults != nil {
			if *request.MaxResults < 1 || *request.MaxResults > 100 {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid max_results value",
					Details: "max_results must be between 1 and 100",
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Validate tag filter if provided
		if request.TagFilter != nil {
			if err := validateTags(request.TagFilter); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    http.StatusBadRequest,
					Message: "Invalid tag filter",
					Details: err.Error(),
				}
				c.JSON(http.StatusBadRequest, errorResponse)
				c.Abort()
				return
			}
		}

		// Store validated request in context for the handler
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// Helper functions

// isValidProjectID checks if a project ID has the correct format
func isValidProjectID(projectID string) bool {
	if len(projectID) < 6 {
		return false
	}

	if !strings.HasPrefix(projectID, "proj-") {
		return false
	}

	// Check if the rest contains only alphanumeric characters and hyphens
	rest := projectID[5:]
	for _, char := range rest {
		if !isValidProjectIDChar(char) {
			return false
		}
	}

	return true
}

// isValidProjectIDChar checks if a character is valid for project ID
func isValidProjectIDChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') || char == '-'
}

// validateTags validates tag key-value pairs
func validateTags(tags map[string]string) error {
	if len(tags) > 10 {
		return fmt.Errorf("too many tags, maximum 10 allowed")
	}

	for key, value := range tags {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("tag key cannot be empty")
		}

		if len(key) > 50 {
			return fmt.Errorf("tag key '%s' too long, maximum 50 characters allowed", key)
		}

		if len(value) > 100 {
			return fmt.Errorf("tag value for key '%s' too long, maximum 100 characters allowed", key)
		}

		// Check for reserved tag prefixes
		if strings.HasPrefix(strings.ToLower(key), "aws:") || strings.HasPrefix(strings.ToLower(key), "system:") {
			return fmt.Errorf("tag key '%s' uses reserved prefix", key)
		}
	}

	return nil
}

// validateMetadata validates metadata key-value pairs
func validateMetadata(metadata map[string]string) error {
	if len(metadata) > 20 {
		return fmt.Errorf("too many metadata entries, maximum 20 allowed")
	}

	for key, value := range metadata {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("metadata key cannot be empty")
		}

		if len(key) > 100 {
			return fmt.Errorf("metadata key '%s' too long, maximum 100 characters allowed", key)
		}

		if len(value) > 500 {
			return fmt.Errorf("metadata value for key '%s' too long, maximum 500 characters allowed", key)
		}
	}

	return nil
}
