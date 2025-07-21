// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
)

// AgentController handles agent-related HTTP requests
type AgentController struct {
	agentService services.AgentService
}

// NewAgentController creates a new AgentController
func NewAgentController(agentService services.AgentService) *AgentController {
	return &AgentController{
		agentService: agentService,
	}
}

// CreateAgent handles POST /agent/create
// @Summary Create a new agent
// @Description Create a new agent for code analysis with the specified repository
// @Tags agents
// @Accept json
// @Produce json
// @Param request body models.CreateAgentRequest true "Agent creation request"
// @Success 201 {object} models.CreateAgentResponse "Agent created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agent/create [post]
func (c *AgentController) CreateAgent(ctx *gin.Context) {
	var request models.CreateAgentRequest

	// Bind and validate the request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Call the service to create the agent
	response, err := c.agentService.CreateAgent(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create agent",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// GetAgent handles GET /agent/:id
// @Summary Get an agent by ID
// @Description Retrieve agent information by agent ID
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} models.CreateAgentResponse "Agent found"
// @Failure 404 {object} models.ErrorResponse "Agent not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agent/{id} [get]
func (c *AgentController) GetAgent(ctx *gin.Context) {
	agentID := ctx.Param("id")
	if agentID == "" {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Agent ID is required",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	response, err := c.agentService.GetAgent(ctx.Request.Context(), agentID)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get agent",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteAgent handles DELETE /agent/:id
// @Summary Delete an agent by ID
// @Description Delete an agent and its associated resources
// @Tags agents
// @Param id path string true "Agent ID"
// @Success 204 "Agent deleted successfully"
// @Failure 404 {object} models.ErrorResponse "Agent not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agent/{id} [delete]
func (c *AgentController) DeleteAgent(ctx *gin.Context) {
	agentID := ctx.Param("id")
	if agentID == "" {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Agent ID is required",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	err := c.agentService.DeleteAgent(ctx.Request.Context(), agentID)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete agent",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListAgents handles GET /agents
// @Summary List all agents
// @Description Get a list of all agents
// @Tags agents
// @Produce json
// @Success 200 {array} models.CreateAgentResponse "List of agents"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents [get]
func (c *AgentController) ListAgents(ctx *gin.Context) {
	responses, err := c.agentService.ListAgents(ctx.Request.Context())
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list agents",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, responses)
}
