package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCloudWatch is a mock implementation of CloudWatch interface
type MockCloudWatch struct {
	mock.Mock
}

// PutMetricData is a mock method for PutMetricData
func (m *MockCloudWatch) PutMetricData(input *cloudwatch.PutMetricDataInput) (*cloudwatch.PutMetricDataOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudwatch.PutMetricDataOutput), args.Error(1)
}

func TestNewMetricsMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		config        config.MetricsConfig
		expectError   bool
		expectEnabled bool
	}{
		{
			name: "enabled_middleware_with_valid_config",
			config: config.MetricsConfig{
				Namespace:   "TestApp/API",
				Region:      "us-west-2",
				ServiceName: "test-service",
				Enabled:     true,
			},
			expectError:   false,
			expectEnabled: true,
		},
		{
			name: "disabled_middleware",
			config: config.MetricsConfig{
				Namespace:   "TestApp/API",
				Region:      "us-west-2",
				ServiceName: "test-service",
				Enabled:     false,
			},
			expectError:   false,
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := NewMetricsMiddleware(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, middleware)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, middleware)
				assert.Equal(t, tt.config, middleware.config)

				if tt.expectEnabled {
					assert.NotNil(t, middleware.cloudWatch)
				} else {
					assert.Nil(t, middleware.cloudWatch)
				}
			}
		})
	}
}

func TestMetricsMiddleware_RequestMetrics_Disabled(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     false,
	}

	middleware, err := NewMetricsMiddleware(config)
	assert.NoError(t, err)

	// Setup test route
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestMetricsMiddleware_RequestMetrics_Success(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     true,
	}

	// Create middleware with disabled CloudWatch for testing
	middleware := &MetricsMiddleware{
		config:     config,
		cloudWatch: nil, // This will prevent actual CloudWatch calls
	}

	// Setup test route
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond) // Simulate some processing time
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestMetricsMiddleware_RequestMetrics_ErrorResponse(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     true,
	}

	// Create middleware with disabled CloudWatch for testing
	middleware := &MetricsMiddleware{
		config:     config,
		cloudWatch: nil,
	}

	// Setup test route that returns an error
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "server error", response["error"])
}

func TestMetricsMiddleware_SendCustomMetric_Disabled(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     false,
	}

	middleware, err := NewMetricsMiddleware(config)
	assert.NoError(t, err)

	// Send custom metric
	err = middleware.SendCustomMetric("TestMetric", 100.0, "Count", map[string]string{
		"Environment": "test",
	})

	// Should not return error even when disabled
	assert.NoError(t, err)
}

func TestMetricsMiddleware_SendCustomMetric_Enabled(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     true,
	}

	// Create middleware with disabled CloudWatch for testing
	middleware := &MetricsMiddleware{
		config:     config,
		cloudWatch: nil,
	}

	// Send custom metric
	err := middleware.SendCustomMetric("TestMetric", 100.0, "Count", map[string]string{
		"Environment": "test",
	})

	// Should not return error even with nil CloudWatch
	assert.NoError(t, err)
}

func TestGetMetricsFromContext(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     false,
	}

	middleware, err := NewMetricsMiddleware(config)
	assert.NoError(t, err)

	// Setup test route
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		// Get metrics from context
		metrics := GetMetricsFromContext(c)
		assert.NotNil(t, metrics)
		assert.Equal(t, middleware, metrics)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMetricsFromContext_NotFound(t *testing.T) {
	// Setup test route without metrics middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		// Get metrics from context
		metrics := GetMetricsFromContext(c)
		assert.Nil(t, metrics)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMetricsFromContext_WrongType(t *testing.T) {
	// Setup test route with wrong type in context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("metrics", "not a metrics middleware")
		c.Next()
	})
	router.GET("/test", func(c *gin.Context) {
		// Get metrics from context
		metrics := GetMetricsFromContext(c)
		assert.Nil(t, metrics)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricsMiddleware_SetMetricsInContext(t *testing.T) {
	config := config.MetricsConfig{
		Namespace:   "TestApp/API",
		Region:      "us-west-2",
		ServiceName: "test-service",
		Enabled:     false,
	}

	middleware, err := NewMetricsMiddleware(config)
	assert.NoError(t, err)

	// Setup test route
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Handle())
	router.GET("/test", func(c *gin.Context) {
		// Verify middleware is in context
		metricsValue, exists := c.Get("metrics")
		assert.True(t, exists)
		assert.Equal(t, middleware, metricsValue)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
}
