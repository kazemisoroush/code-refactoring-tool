package middleware

import (
	"fmt"
	"reflect"
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

// ValidationMiddleware provides validation middleware for request data
type ValidationMiddleware struct {
	validationType string
	validatorFunc  gin.HandlerFunc
}

// ValidationConfig defines the configuration for validation middleware
type ValidationConfig struct {
	Type          string // "json", "uri", "query", "combined"
	ValidatorFunc gin.HandlerFunc
}

// NewValidationMiddleware creates a new validation middleware with the specified config
func NewValidationMiddleware(config ValidationConfig) Middleware {
	return &ValidationMiddleware{
		validationType: config.Type,
		validatorFunc:  config.ValidatorFunc,
	}
}

// Handle implements the Middleware interface and returns the validation handler
func (m *ValidationMiddleware) Handle() gin.HandlerFunc {
	return m.validatorFunc
}

// Factory methods for creating specific validation middlewares

// NewJSONValidationMiddleware creates a JSON validation middleware for type T
func NewJSONValidationMiddleware[T any]() Middleware {
	return NewValidationMiddleware(ValidationConfig{
		Type: "json",
		ValidatorFunc: func(c *gin.Context) {
			var request T
			if err := c.ShouldBind(&request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Invalid request body",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			if err := validate.Struct(request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Validation failed",
					Details: formatValidationError(err),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			c.Set("validatedRequest", request)
			c.Next()
		},
	})
}

// NewURIValidationMiddleware creates a URI validation middleware for type T
func NewURIValidationMiddleware[T any]() Middleware {
	return NewValidationMiddleware(ValidationConfig{
		Type: "uri",
		ValidatorFunc: func(c *gin.Context) {
			var request T
			if err := c.ShouldBindUri(&request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Invalid URI parameters",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			if err := validate.Struct(request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Validation failed",
					Details: formatValidationError(err),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			c.Set("validatedRequest", request)
			c.Next()
		},
	})
}

// NewQueryValidationMiddleware creates a query validation middleware for type T
func NewQueryValidationMiddleware[T any]() Middleware {
	return NewValidationMiddleware(ValidationConfig{
		Type: "query",
		ValidatorFunc: func(c *gin.Context) {
			var request T
			if err := c.ShouldBindQuery(&request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Invalid query parameters",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			if err := validate.Struct(request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Validation failed",
					Details: formatValidationError(err),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			c.Set("validatedRequest", request)
			c.Next()
		},
	})
}

// NewCombinedValidationMiddleware creates a combined validation middleware for type T
func NewCombinedValidationMiddleware[T any]() Middleware {
	return NewValidationMiddleware(ValidationConfig{
		Type: "combined",
		ValidatorFunc: func(c *gin.Context) {
			var request T
			if err := c.ShouldBindUri(&request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Invalid URI parameters",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			var jsonRequest T
			if err := c.ShouldBindJSON(&jsonRequest); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Invalid request body",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			if err := mergeStructs(&request, jsonRequest); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Failed to merge request data",
					Details: err.Error(),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			if err := validate.Struct(request); err != nil {
				errorResponse := models.ErrorResponse{
					Code:    400,
					Message: "Validation failed",
					Details: formatValidationError(err),
				}
				c.JSON(400, errorResponse)
				c.Abort()
				return
			}
			c.Set("validatedRequest", request)
			c.Next()
		},
	})
}

// Custom validation functions

// validateProjectID validates project ID format - just ensures it's a non-empty string
func validateProjectID(fl validator.FieldLevel) bool {
	projectID := fl.Field().String()
	return len(projectID) > 0
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
