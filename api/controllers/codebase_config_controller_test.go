package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	servicesMocks "github.com/kazemisoroush/code-refactoring-tool/api/services/mocks"
)

func TestCodebaseConfigController_CreateCodebaseConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockCodebaseConfigService(ctrl)
	controller := controllers.NewCodebaseConfigController(mockService)

	request := models.CreateCodebaseConfigRequest{
		Name:     "My Config",
		Provider: "github",
		URL:      "https://github.com/test/repo.git",
		Config: models.GitProviderConfig{
			AuthType: "token",
			GitHub: &models.GitHubConfig{
				Token:         "test-token",
				Owner:         "test-owner",
				Repository:    "test-repo",
				DefaultBranch: "main",
			},
		},
	}

	expectedResponse := &models.CreateCodebaseConfigResponse{
		ConfigID:  "config-123",
		CreatedAt: "2025-08-10T10:00:00Z",
	}

	mockService.EXPECT().CreateCodebaseConfig(gomock.Any(), request).Return(expectedResponse, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/codebase-configs", func(ctx *gin.Context) {
		// Simulate validation middleware
		ctx.Set("validatedRequest", request)
		controller.CreateCodebaseConfig(ctx)
	})

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/codebase-configs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response models.CreateCodebaseConfigResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ConfigID, response.ConfigID)
	assert.Equal(t, expectedResponse.CreatedAt, response.CreatedAt)
}

func TestCodebaseConfigController_GetCodebaseConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockCodebaseConfigService(ctrl)
	controller := controllers.NewCodebaseConfigController(mockService)

	configID := "config-123"
	request := models.GetCodebaseConfigRequest{ConfigID: configID}

	expectedResponse := &models.GetCodebaseConfigResponse{
		ConfigID:  configID,
		Name:      "Test Config",
		Provider:  "github",
		URL:       "https://github.com/test/repo.git",
		CreatedAt: "2025-08-10T10:00:00Z",
		UpdatedAt: "2025-08-10T10:00:00Z",
	}

	mockService.EXPECT().GetCodebaseConfig(gomock.Any(), configID).Return(expectedResponse, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/codebase-configs/:config_id", func(ctx *gin.Context) {
		ctx.Set("validatedRequest", request)
		controller.GetCodebaseConfig(ctx)
	})

	req := httptest.NewRequest(http.MethodGet, "/codebase-configs/"+configID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.GetCodebaseConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ConfigID, response.ConfigID)
	assert.Equal(t, expectedResponse.Name, response.Name)
}

func TestCodebaseConfigController_UpdateCodebaseConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockCodebaseConfigService(ctrl)
	controller := controllers.NewCodebaseConfigController(mockService)

	configID := "config-123"
	name := "Updated Config"
	url := "https://github.com/test/updated-repo.git"
	request := models.UpdateCodebaseConfigRequest{
		ConfigID: configID,
		Name:     &name,
		URL:      &url,
		Config:   &models.GitProviderConfig{},
	}

	expectedResponse := &models.UpdateCodebaseConfigResponse{
		ConfigID:  configID,
		UpdatedAt: "2025-08-10T10:30:00Z",
	}

	mockService.EXPECT().UpdateCodebaseConfig(gomock.Any(), request).Return(expectedResponse, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/codebase-configs/:config_id", func(ctx *gin.Context) {
		ctx.Set("validatedRequest", request)
		controller.UpdateCodebaseConfig(ctx)
	})

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/codebase-configs/"+configID, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.UpdateCodebaseConfigResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.ConfigID, response.ConfigID)
	assert.Equal(t, expectedResponse.UpdatedAt, response.UpdatedAt)
}

func TestCodebaseConfigController_DeleteCodebaseConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockCodebaseConfigService(ctrl)
	controller := controllers.NewCodebaseConfigController(mockService)

	configID := "config-123"
	request := models.DeleteCodebaseConfigRequest{ConfigID: configID}

	expectedResponse := &models.DeleteCodebaseConfigResponse{
		Success: true,
	}

	mockService.EXPECT().DeleteCodebaseConfig(gomock.Any(), configID).Return(expectedResponse, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/codebase-configs/:config_id", func(ctx *gin.Context) {
		ctx.Set("validatedRequest", request)
		controller.DeleteCodebaseConfig(ctx)
	})

	req := httptest.NewRequest(http.MethodDelete, "/codebase-configs/"+configID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.DeleteCodebaseConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse.Success, response.Success)
}

func TestCodebaseConfigController_ListCodebaseConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockCodebaseConfigService(ctrl)
	controller := controllers.NewCodebaseConfigController(mockService)

	maxResults := 10
	request := models.ListCodebaseConfigsRequest{
		MaxResults: &maxResults,
	}

	nextToken := ""
	expectedResponse := &models.ListCodebaseConfigsResponse{
		Configs: []models.CodebaseConfigSummary{
			{
				ConfigID:  "config-123",
				Name:      "Test Config 1",
				Provider:  "github",
				URL:       "https://github.com/test/repo1.git",
				CreatedAt: "2025-08-10T10:00:00Z",
			},
			{
				ConfigID:  "config-456",
				Name:      "Test Config 2",
				Provider:  "gitlab",
				URL:       "https://gitlab.com/test/repo2.git",
				CreatedAt: "2025-08-10T11:00:00Z",
			},
		},
		NextToken: &nextToken,
	}

	mockService.EXPECT().ListCodebaseConfigs(gomock.Any(), request).Return(expectedResponse, nil).Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/codebase-configs", func(ctx *gin.Context) {
		ctx.Set("validatedRequest", request)
		controller.ListCodebaseConfigs(ctx)
	})

	req := httptest.NewRequest(http.MethodGet, "/codebase-configs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response models.ListCodebaseConfigsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Configs, 2)
	assert.Equal(t, expectedResponse.Configs[0].ConfigID, response.Configs[0].ConfigID)
	assert.Equal(t, expectedResponse.Configs[1].ConfigID, response.Configs[1].ConfigID)
}
