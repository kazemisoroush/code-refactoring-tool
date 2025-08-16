package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services/mocks"
)

func TestAuthController_ConfirmEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{} // Using interface{} to allow invalid JSON
		setupMock      func(*mocks.MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful email confirmation",
			requestBody: models.ConfirmEmailRequest{
				Email: "testuser@example.com",
				Code:  "123456",
			},
			setupMock: func(mockAuthService *mocks.MockAuthService) {
				mockAuthService.EXPECT().
					ConfirmEmail(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Account verified successfully"}`,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "", // empty required field
				"code":  "123456",
			},
			setupMock:      func(*mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			requestBody: models.ConfirmEmailRequest{
				Email: "nonexistent@example.com",
				Code:  "123456",
			},
			setupMock: func(mockAuthService *mocks.MockAuthService) {
				mockAuthService.EXPECT().
					ConfirmEmail(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthService(ctrl)
			authController := NewAuthController(mockAuthService)

			tt.setupMock(mockAuthService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Use validation middleware like in the actual routes
			router.POST("/auth/confirm", middleware.NewJSONValidationMiddleware[models.ConfirmEmailRequest]().Handle(), authController.ConfirmEmail)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/confirm", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthController_ForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{} // Using interface{} to allow invalid JSON
		setupMock      func(*mocks.MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful password reset initiation",
			requestBody: models.ForgotPasswordRequest{
				Email: "testuser@example.com",
			},
			setupMock: func(mockAuthService *mocks.MockAuthService) {
				mockAuthService.EXPECT().
					ForgotPassword(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Password reset code sent to your email"}`,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"email": "", // empty required field
			},
			setupMock:      func(*mocks.MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthService := mocks.NewMockAuthService(ctrl)
			authController := NewAuthController(mockAuthService)

			tt.setupMock(mockAuthService)

			gin.SetMode(gin.TestMode)
			router := gin.New()

			// Use validation middleware like in the actual routes
			router.POST("/auth/forgot-password", middleware.NewJSONValidationMiddleware[models.ForgotPasswordRequest]().Handle(), authController.ForgotPassword)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/forgot-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestAuthController_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	authController := NewAuthController(mockAuthService)

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:   "successful get current user",
			userID: "user123",
			setupMock: func() {
				expectedResponse := &models.GetUserResponse{
					User: &models.APIUser{
						UserID:   "user123",
						Username: "testuser",
						Email:    "test@example.com",
					},
				}
				mockAuthService.EXPECT().
					GetUser(gomock.Any(), "user123").
					Return(expectedResponse, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not authenticated",
			userID:         nil,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid user ID type",
			userID:         123, // not a string
			setupMock:      func() {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userID != nil {
				c.Set("userID", tt.userID)
			}

			c.Request = httptest.NewRequest("GET", "/auth/me", nil)

			authController.GetMe(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
