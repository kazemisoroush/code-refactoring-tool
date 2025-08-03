// Package middleware provides HTTP middleware components for the API
package middleware

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// MetricsMiddleware provides middleware for collecting and sending metrics to CloudWatch
type MetricsMiddleware struct {
	config     config.MetricsConfig
	cloudWatch *cloudwatch.CloudWatch
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(config config.MetricsConfig) (*MetricsMiddleware, error) {
	if !config.Enabled {
		return &MetricsMiddleware{config: config}, nil
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		return nil, err
	}

	// Create CloudWatch client
	cw := cloudwatch.New(sess)

	return &MetricsMiddleware{
		config:     config,
		cloudWatch: cw,
	}, nil
}

// Handle is the middleware function that collects HTTP request metrics and sets context
func (m *MetricsMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set metrics middleware in context for downstream handlers
		c.Set("metrics", m)

		if !m.config.Enabled {
			c.Next()
			return
		}

		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate metrics
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Send metrics asynchronously to avoid blocking the request
		go m.sendRequestMetrics(method, path, statusCode, duration)
	}
}

// sendRequestMetrics sends HTTP request metrics to CloudWatch
func (m *MetricsMiddleware) sendRequestMetrics(method, path string, statusCode int, duration time.Duration) {
	if m.cloudWatch == nil {
		return
	}

	timestamp := time.Now()

	// Prepare metrics data
	metricData := []*cloudwatch.MetricDatum{
		// Request count metric
		{
			MetricName: aws.String("RequestCount"),
			Dimensions: []*cloudwatch.Dimension{
				{
					Name:  aws.String("Method"),
					Value: aws.String(method),
				},
				{
					Name:  aws.String("Path"),
					Value: aws.String(path),
				},
				{
					Name:  aws.String("StatusCode"),
					Value: aws.String(strconv.Itoa(statusCode)),
				},
				{
					Name:  aws.String("Service"),
					Value: aws.String(m.config.ServiceName),
				},
			},
			Timestamp: aws.Time(timestamp),
			Unit:      aws.String("Count"),
			Value:     aws.Float64(1),
		},
		// Response time metric
		{
			MetricName: aws.String("ResponseTime"),
			Dimensions: []*cloudwatch.Dimension{
				{
					Name:  aws.String("Method"),
					Value: aws.String(method),
				},
				{
					Name:  aws.String("Path"),
					Value: aws.String(path),
				},
				{
					Name:  aws.String("Service"),
					Value: aws.String(m.config.ServiceName),
				},
			},
			Timestamp: aws.Time(timestamp),
			Unit:      aws.String("Milliseconds"),
			Value:     aws.Float64(float64(duration.Nanoseconds()) / 1e6),
		},
	}

	// Add error metric for 4xx and 5xx responses
	if statusCode >= 400 {
		errorType := "ClientError"
		if statusCode >= 500 {
			errorType = "ServerError"
		}

		errorMetric := &cloudwatch.MetricDatum{
			MetricName: aws.String("ErrorCount"),
			Dimensions: []*cloudwatch.Dimension{
				{
					Name:  aws.String("ErrorType"),
					Value: aws.String(errorType),
				},
				{
					Name:  aws.String("Method"),
					Value: aws.String(method),
				},
				{
					Name:  aws.String("Path"),
					Value: aws.String(path),
				},
				{
					Name:  aws.String("StatusCode"),
					Value: aws.String(strconv.Itoa(statusCode)),
				},
				{
					Name:  aws.String("Service"),
					Value: aws.String(m.config.ServiceName),
				},
			},
			Timestamp: aws.Time(timestamp),
			Unit:      aws.String("Count"),
			Value:     aws.Float64(1),
		}
		metricData = append(metricData, errorMetric)
	}

	// Send metrics to CloudWatch
	_, err := m.cloudWatch.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(m.config.Namespace),
		MetricData: metricData,
	})

	if err != nil {
		// Log error but don't fail the request
		// In a production environment, you might want to use a proper logger here
		// For now, we'll silently ignore errors to avoid affecting request processing
		_ = err
	}
}

// SendCustomMetric sends a custom metric to CloudWatch
func (m *MetricsMiddleware) SendCustomMetric(metricName string, value float64, unit string, dimensions map[string]string) error {
	if !m.config.Enabled || m.cloudWatch == nil {
		return nil
	}

	// Convert dimensions map to CloudWatch dimensions
	var cwDimensions []*cloudwatch.Dimension
	for name, value := range dimensions {
		cwDimensions = append(cwDimensions, &cloudwatch.Dimension{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	// Add service dimension
	cwDimensions = append(cwDimensions, &cloudwatch.Dimension{
		Name:  aws.String("Service"),
		Value: aws.String(m.config.ServiceName),
	})

	metricData := []*cloudwatch.MetricDatum{
		{
			MetricName: aws.String(metricName),
			Dimensions: cwDimensions,
			Timestamp:  aws.Time(time.Now()),
			Unit:       aws.String(unit),
			Value:      aws.Float64(value),
		},
	}

	_, err := m.cloudWatch.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace:  aws.String(m.config.Namespace),
		MetricData: metricData,
	})

	return err
}

// GetMetricsFromContext retrieves the metrics middleware from gin context
func GetMetricsFromContext(c *gin.Context) *MetricsMiddleware {
	if metrics, exists := c.Get("metrics"); exists {
		if m, ok := metrics.(*MetricsMiddleware); ok {
			return m
		}
	}
	return nil
}
