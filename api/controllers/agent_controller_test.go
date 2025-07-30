package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	servicesMocks "github.com/kazemisoroush/code-refactoring-tool/api/services/mocks"
)

func TestNewAgentController(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	assert.NotNil(t, controller)
	assert.Equal(t, mockService, controller.agentService)
}

func TestAgentController_CreateAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	request := models.CreateAgentRequest{
		RepositoryURL: "https://github.com/user/repo",
		Branch:        "main",
		AgentName:     "test-agent",
	}

	expectedResponse := &models.CreateAgentResponse{
		AgentID:         "agent-123",
		AgentVersion:    "v1.0.0",
		KnowledgeBaseID: "kb-456",
		VectorStoreID:   "vs-789",
		Status:          string(models.AgentStatusReady),
		CreatedAt:       time.Now().UTC(),
	}

	mockService.EXPECT().
		CreateAgent(gomock.Any(), request).
		Return(expectedResponse, nil).
		Times(1)

	// Create HTTP request
	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/agent/create", controller.CreateAgent)

	req := httptest.NewRequest(http.MethodPost, "/agent/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.CreateAgentResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.AgentID, response.AgentID)
	assert.Equal(t, expectedResponse.AgentVersion, response.AgentVersion)
	assert.Equal(t, expectedResponse.KnowledgeBaseID, response.KnowledgeBaseID)
	assert.Equal(t, expectedResponse.VectorStoreID, response.VectorStoreID)
}

func TestAgentController_CreateAgent_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	// Invalid JSON
	invalidJSON := `{"repository_url":}`

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/agent/create", controller.CreateAgent)

	req := httptest.NewRequest(http.MethodPost, "/agent/create", bytes.NewReader([]byte(invalidJSON)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
	assert.Equal(t, "Invalid request body", errorResponse.Message)
}

func TestAgentController_CreateAgent_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	request := models.CreateAgentRequest{
		RepositoryURL: "https://github.com/user/repo",
	}

	serviceError := errors.New("service error")
	mockService.EXPECT().
		CreateAgent(gomock.Any(), request).
		Return(nil, serviceError).
		Times(1)

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/agent/create", controller.CreateAgent)

	req := httptest.NewRequest(http.MethodPost, "/agent/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
	assert.Equal(t, "Failed to create agent", errorResponse.Message)
}

func TestAgentController_GetAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	agentID := "agent-123"
	expectedResponse := &models.GetAgentResponse{
		AgentID:         agentID,
		AgentVersion:    "v1.0.0",
		KnowledgeBaseID: "kb-456",
		VectorStoreID:   "vs-789",
		RepositoryURL:   "https://github.com/test/repo",
		Branch:          "main",
		AgentName:       "test-agent",
		Status:          string(models.AgentStatusReady),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	mockService.EXPECT().
		GetAgent(gomock.Any(), agentID).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/agent/:id", controller.GetAgent)

	req := httptest.NewRequest(http.MethodGet, "/agent/"+agentID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetAgentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.AgentID, response.AgentID)
}

func TestAgentController_DeleteAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	agentID := "agent-123"
	expectedResponse := &models.DeleteAgentResponse{
		AgentID: agentID,
		Success: true,
	}

	mockService.EXPECT().
		DeleteAgent(gomock.Any(), agentID).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/agent/:id", controller.DeleteAgent)

	req := httptest.NewRequest(http.MethodDelete, "/agent/"+agentID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeleteAgentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.AgentID, response.AgentID)
	assert.Equal(t, expectedResponse.Success, response.Success)
}

func TestAgentController_ListAgents_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := servicesMocks.NewMockAgentService(ctrl)
	controller := NewAgentController(mockService)

	expectedResponse := &models.ListAgentsResponse{
		Agents: []models.AgentSummary{
			{
				AgentID:       "agent-123",
				AgentName:     "test-agent",
				RepositoryURL: "https://github.com/test/repo",
				Status:        string(models.AgentStatusReady),
				CreatedAt:     time.Now().UTC(),
			},
		},
	}

	mockService.EXPECT().
		ListAgents(gomock.Any(), gomock.Any()).
		Return(expectedResponse, nil).
		Times(1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/agents", controller.ListAgents)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.ListAgentsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Agents, 1)
	assert.Equal(t, expectedResponse.Agents[0].AgentID, response.Agents[0].AgentID)
}
