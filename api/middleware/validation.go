package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()

	// Register custom validations
	_ = validate.RegisterValidation("project_id", validateProjectID)
	_ = validate.RegisterValidation("provider", validateProvider)
	_ = validate.RegisterValidation("tag_filter", validateTagFilter)

	// Register custom tag name function for better error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		}
		if name == "" {
			name = strings.SplitN(fld.Tag.Get("uri"), ",", 2)[0]
		}
		// Special case for common field names
		switch name {
		case "id":
			return "ID"
		case "max_results":
			return "MaxResults"
		case "next_token":
			return "NextToken"
		case "tag_filter":
			return "TagFilter"
		default:
			return name
		}
	})
}

// ValidateJSON validates JSON request body using struct tags
func ValidateJSON[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request T

		// Use ShouldBind instead of ShouldBindJSON to skip Gin's validation
		if err := c.ShouldBind(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		if err := validate.Struct(request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Validation failed",
				Details: formatValidationError(err),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// ValidateURI validates URI parameters using struct tags
func ValidateURI[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request T

		if err := c.ShouldBindUri(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid URI parameters",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		if err := validate.Struct(request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Validation failed",
				Details: formatValidationError(err),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// ValidateQuery validates query parameters using struct tags
func ValidateQuery[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request T

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

		if err := validate.Struct(request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Validation failed",
				Details: formatValidationError(err),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// ValidateCombined validates both URI parameters and JSON body
func ValidateCombined[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request T

		// Bind URI parameters first
		if err := c.ShouldBindUri(&request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid URI parameters",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Create a temporary struct for JSON to preserve URI fields
		var jsonRequest T
		if err := c.ShouldBindJSON(&jsonRequest); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Merge the two structs - URI fields take precedence
		if err := mergeStructs(&request, jsonRequest); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Failed to merge request data",
				Details: err.Error(),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		if err := validate.Struct(request); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Validation failed",
				Details: formatValidationError(err),
			}
			c.JSON(http.StatusBadRequest, errorResponse)
			c.Abort()
			return
		}

		// Store validated request in context
		c.Set("validatedRequest", request)
		c.Next()
	}
}

// Custom validation functions

// validateProjectID validates project ID format
func validateProjectID(fl validator.FieldLevel) bool {
	projectID := fl.Field().String()

	if len(projectID) < 6 {
		return false
	}

	if !strings.HasPrefix(projectID, "proj-") {
		return false
	}

	// Check if the rest contains only alphanumeric characters and hyphens
	rest := projectID[5:]
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, rest)
	return matched
}

// validateProvider validates that the provider is one of the supported types
func validateProvider(fl validator.FieldLevel) bool {
	provider := models.Provider(fl.Field().String())
	return provider.IsValid()
}

// validateTagFilter validates tag filter format (key:value)
func validateTagFilter(fl validator.FieldLevel) bool {
	tagFilter := fl.Field().String()
	if tagFilter == "" {
		return true // optional field
	}

	// Check format: key:value
	parts := strings.SplitN(tagFilter, ":", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// formatValidationError formats validation errors into human-readable messages
func formatValidationError(err error) string {
	var messages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			message := formatFieldError(fieldError)
			messages = append(messages, message)
		}
	}

	if len(messages) == 0 {
		return err.Error()
	}

	return strings.Join(messages, "; ")
}

// formatFieldError formats a single field validation error
func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	if field == "" {
		field = fe.Tag()
	}

	// Capitalize the first letter of the field name for better user experience
	if len(field) > 0 {
		field = strings.ToUpper(field[:1]) + field[1:]
	}

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "oneof":
		options := strings.Split(fe.Param(), " ")
		return fmt.Sprintf("%s must be one of %v", field, options)
	case "project_id":
		return fmt.Sprintf("%s must start with 'proj-' followed by alphanumeric characters", field)
	case "provider":
		return fmt.Sprintf("%s must be one of: github, gitlab, bitbucket, custom", field)
	case "tag_filter":
		return fmt.Sprintf("%s must be in format key:value", field)
	case "dive":
		return fmt.Sprintf("%s contains invalid values", field)
	case "keys":
		return fmt.Sprintf("%s contains invalid keys", field)
	case "endkeys":
		return fmt.Sprintf("%s contains invalid key-value pairs", field)
	default:
		return fmt.Sprintf("%s is invalid: %s", field, fe.Tag())
	}
}

// GetValidatedRequest retrieves the validated request from gin context
func GetValidatedRequest[T any](c *gin.Context) (T, bool) {
	var zero T
	value, exists := c.Get("validatedRequest")
	if !exists {
		return zero, false
	}

	if typed, ok := value.(T); ok {
		return typed, true
	}

	return zero, false
}

// mergeStructs merges JSON fields into URI struct, preserving URI fields
func mergeStructs(dest interface{}, src interface{}) error {
	destVal := reflect.ValueOf(dest).Elem()
	srcVal := reflect.ValueOf(src)

	destType := destVal.Type()
	srcType := srcVal.Type()

	if destType != srcType {
		return fmt.Errorf("types do not match")
	}

	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		srcField := srcVal.Field(i)
		fieldType := destType.Field(i)

		// Skip fields that have URI tags (preserve URI-bound values)
		if fieldType.Tag.Get("uri") != "" {
			continue
		}

		// Only copy non-zero values from JSON
		if srcField.IsValid() && !srcField.IsZero() && destField.CanSet() {
			destField.Set(srcField)
		}
	}

	return nil
}
