package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test models for validation
type TestCreateRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=100" example:"test-name"`
	Email       string            `json:"email" validate:"required,email" example:"test@example.com"`
	Age         *int              `json:"age,omitempty" validate:"omitempty,min=0,max=150" example:"25"`
	Language    *string           `json:"language,omitempty" validate:"omitempty,oneof=go python java javascript typescript" example:"go"`
	Tags        map[string]string `json:"tags,omitempty" validate:"omitempty,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=500" example:"Test description"`
}

type TestUpdateRequest struct {
	ID          string            `uri:"id" validate:"required,project_id" example:"proj-12345"`
	Name        *string           `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"updated-name"`
	Email       *string           `json:"email,omitempty" validate:"omitempty,email" example:"updated@example.com"`
	Tags        map[string]string `json:"tags,omitempty" validate:"omitempty,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:staging"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated description"`
}

type TestListRequest struct {
	MaxResults *int              `form:"max_results,omitempty" validate:"omitempty,min=1,max=100" example:"50"`
	NextToken  *string           `form:"next_token,omitempty" validate:"omitempty,min=1" example:"token123"`
	TagFilter  map[string]string `form:"tag_filter,omitempty" validate:"omitempty,dive,keys,min=1,max=50,endkeys,min=1,max=100" example:"env:prod"`
}

func TestValidateJSON_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		request TestCreateRequest
	}{
		{
			name: "valid complete request",
			request: TestCreateRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				Age:         intPtr(25),
				Language:    stringPtr("go"),
				Tags:        map[string]string{"env": "prod", "team": "backend"},
				Description: stringPtr("A test user"),
			},
		},
		{
			name: "valid minimal request",
			request: TestCreateRequest{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute
			middleware := NewJSONValidationMiddleware[TestCreateRequest]().Handle()
			middleware(c)

			// Assert
			assert.False(t, c.IsAborted())

			// Check that validated request is stored in context
			validatedRequest, exists := c.Get("validatedRequest")
			assert.True(t, exists)
			assert.IsType(t, TestCreateRequest{}, validatedRequest)
		})
	}
}

func TestValidateJSON_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        TestCreateRequest
		expectedError  string
		expectedStatus int
	}{
		{
			name: "missing required name",
			request: TestCreateRequest{
				Email: "john@example.com",
			},
			expectedError:  "Name is required",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			request: TestCreateRequest{
				Name:  "John Doe",
				Email: "invalid-email",
			},
			expectedError:  "Email must be a valid email address",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "age out of range",
			request: TestCreateRequest{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   intPtr(200),
			},
			expectedError:  "Age must be at most 150",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid language",
			request: TestCreateRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Language: stringPtr("cobol"),
			},
			expectedError:  "Language must be one of [go python java javascript typescript]",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			request: TestCreateRequest{
				Name:  "This is a very long name that exceeds the maximum allowed length of 100 characters and should fail validation",
				Email: "john@example.com",
			},
			expectedError:  "Name must be at most 100 characters",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "description too long",
			request: TestCreateRequest{
				Name:        "John Doe",
				Email:       "john@example.com",
				Description: stringPtr(string(make([]byte, 501))), // 501 characters
			},
			expectedError:  "Description must be at most 500 characters",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
			c.Request.Header.Set("Content-Type", "application/json")

			// Execute
			middleware := NewJSONValidationMiddleware[TestCreateRequest]().Handle()
			middleware(c)

			// Assert
			assert.True(t, c.IsAborted())
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check if the error message is in message or details field
			errorMessage := ""
			if msg, ok := response["message"].(string); ok {
				errorMessage += msg
			}
			if details, ok := response["details"].(string); ok {
				errorMessage += " " + details
			}

			assert.Contains(t, errorMessage, tt.expectedError)
		})
	}
}

func TestValidateURI_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("PUT", "/test/proj-12345", nil)
	c.Params = []gin.Param{{Key: "id", Value: "proj-12345"}}

	// Execute
	middleware := NewURIValidationMiddleware[TestUpdateRequest]().Handle()
	middleware(c)

	// Assert
	assert.False(t, c.IsAborted())

	// Check that validated request is stored in context
	validatedRequest, exists := c.Get("validatedRequest")
	assert.True(t, exists)

	req := validatedRequest.(TestUpdateRequest)
	assert.Equal(t, "proj-12345", req.ID)
}

func TestValidateURI_InvalidProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		projectID      string
		expectedError  string
		expectedStatus int
	}{
		{
			name:           "invalid project ID format",
			projectID:      "invalid-id",
			expectedError:  "ID must start with 'proj-' followed by alphanumeric characters",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty project ID",
			projectID:      "",
			expectedError:  "ID is required",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("PUT", "/test/"+tt.projectID, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.projectID}}

			// Execute
			middleware := NewURIValidationMiddleware[TestUpdateRequest]().Handle()
			middleware(c)

			// Assert
			assert.True(t, c.IsAborted())
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check if the error message is in message or details field
			errorMessage := ""
			if msg, ok := response["message"].(string); ok {
				errorMessage += msg
			}
			if details, ok := response["details"].(string); ok {
				errorMessage += " " + details
			}

			assert.Contains(t, errorMessage, tt.expectedError)
		})
	}
}

func TestValidateQuery_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryParams map[string]string
		expected    TestListRequest
	}{
		{
			name: "valid complete query",
			queryParams: map[string]string{
				"max_results": "50",
				"next_token":  "token123",
			},
			expected: TestListRequest{
				MaxResults: intPtr(50),
				NextToken:  stringPtr("token123"),
			},
		},
		{
			name:        "empty query",
			queryParams: map[string]string{},
			expected:    TestListRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build query string
			values := url.Values{}
			for k, v := range tt.queryParams {
				values.Set(k, v)
			}

			c.Request = httptest.NewRequest("GET", "/test?"+values.Encode(), nil)

			// Execute
			middleware := NewQueryValidationMiddleware[TestListRequest]().Handle()
			middleware(c)

			// Assert
			assert.False(t, c.IsAborted())

			// Check that validated request is stored in context
			validatedRequest, exists := c.Get("validatedRequest")
			assert.True(t, exists)

			req := validatedRequest.(TestListRequest)
			if tt.expected.MaxResults != nil {
				assert.Equal(t, *tt.expected.MaxResults, *req.MaxResults)
			}
			if tt.expected.NextToken != nil {
				assert.Equal(t, *tt.expected.NextToken, *req.NextToken)
			}
		})
	}
}

func TestValidateQuery_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedError  string
		expectedStatus int
	}{
		{
			name: "max_results out of range - too high",
			queryParams: map[string]string{
				"max_results": "200",
			},
			expectedError:  "MaxResults must be at most 100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "max_results out of range - too low",
			queryParams: map[string]string{
				"max_results": "0",
			},
			expectedError:  "MaxResults must be at least 1",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build query string
			values := url.Values{}
			for k, v := range tt.queryParams {
				values.Set(k, v)
			}

			c.Request = httptest.NewRequest("GET", "/test?"+values.Encode(), nil)

			// Execute
			middleware := NewQueryValidationMiddleware[TestListRequest]().Handle()
			middleware(c)

			// Assert
			assert.True(t, c.IsAborted())
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Check if the error message is in message or details field
			errorMessage := ""
			if msg, ok := response["message"].(string); ok {
				errorMessage += msg
			}
			if details, ok := response["details"].(string); ok {
				errorMessage += " " + details
			}

			assert.Contains(t, errorMessage, tt.expectedError)
		})
	}
}

func TestValidateCombined_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create request body
	request := TestUpdateRequest{
		Name:        stringPtr("Updated Name"),
		Email:       stringPtr("updated@example.com"),
		Tags:        map[string]string{"env": "staging"},
		Description: stringPtr("Updated description"),
	}
	jsonData, err := json.Marshal(request)
	require.NoError(t, err)

	c.Request = httptest.NewRequest("PUT", "/test/proj-12345", bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "proj-12345"}}

	// Execute combined validation
	middleware := NewCombinedValidationMiddleware[TestUpdateRequest]().Handle()
	middleware(c)

	// Assert
	assert.False(t, c.IsAborted())

	// Check that validated request is stored in context
	validatedRequest, exists := c.Get("validatedRequest")
	assert.True(t, exists)

	req := validatedRequest.(TestUpdateRequest)
	assert.Equal(t, "proj-12345", req.ID)
	assert.Equal(t, "Updated Name", *req.Name)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
