// Package controllers provides HTTP request handlers for the API
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
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

// CreateAgent handles POST /agents
// @Summary Create a new agent
// @Description Create a new agent for code analysis with the specified repository
// @Tags agents
// @Accept json
// @Produce json
// @Param request body models.CreateAgentRequest true "Agent creation request"
// @Success 201 {object} models.CreateAgentResponse "Agent created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents [post]
func (c *AgentController) CreateAgent(ctx *gin.Context) {
	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.CreateAgentRequest](ctx)
	if !exists {
		// Fall back to manual binding for backward compatibility
		var requestData models.CreateAgentRequest
		if err := ctx.ShouldBindJSON(&requestData); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			ctx.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		request = requestData
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

// GetAgent handles GET /agents/:id
// @Summary Get an agent by ID
// @Description Retrieve agent information by agent ID
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} models.GetAgentResponse "Agent found"
// @Failure 400 {object} models.ErrorResponse "Invalid agent ID"
// @Failure 404 {object} models.ErrorResponse "Agent not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents/{id} [get]
func (c *AgentController) GetAgent(ctx *gin.Context) {
	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.GetAgentRequest](ctx)
	var agentID string
	if !exists {
		// Fall back to manual parameter extraction for backward compatibility
		agentID = ctx.Param("id")
		if agentID == "" {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Agent ID is required",
			}
			ctx.JSON(http.StatusBadRequest, errorResponse)
			return
		}
	} else {
		agentID = request.AgentID
	}

	response, err := c.agentService.GetAgent(ctx.Request.Context(), agentID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "agent not found" {
			statusCode = http.StatusNotFound
			message = "Agent not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to get agent"
		}

		errorResponse := models.ErrorResponse{
			Code:    statusCode,
			Message: message,
			Details: err.Error(),
		}
		ctx.JSON(statusCode, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteAgent handles DELETE /agents/:id
// @Summary Delete an agent by ID
// @Description Delete an agent and its associated resources
// @Tags agents
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} models.DeleteAgentResponse "Agent deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid agent ID"
// @Failure 404 {object} models.ErrorResponse "Agent not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents/{id} [delete]
func (c *AgentController) DeleteAgent(ctx *gin.Context) {
	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.DeleteAgentRequest](ctx)
	var agentID string
	if !exists {
		// Fall back to manual parameter extraction for backward compatibility
		agentID = ctx.Param("id")
		if agentID == "" {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Agent ID is required",
			}
			ctx.JSON(http.StatusBadRequest, errorResponse)
			return
		}
	} else {
		agentID = request.AgentID
	}

	response, err := c.agentService.DeleteAgent(ctx.Request.Context(), agentID)
	if err != nil {
		var statusCode int
		var message string

		// Check if it's a "not found" error
		if err.Error() == "agent not found" {
			statusCode = http.StatusNotFound
			message = "Agent not found"
		} else {
			statusCode = http.StatusInternalServerError
			message = "Failed to delete agent"
		}

		errorResponse := models.ErrorResponse{
			Code:    statusCode,
			Message: message,
			Details: err.Error(),
		}
		ctx.JSON(statusCode, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ListAgents handles GET /agents
// @Summary List all agents
// @Description Get a list of agents with optional pagination
// @Tags agents
// @Produce json
// @Param next_token query string false "Token for pagination"
// @Param max_results query int false "Maximum number of results to return" minimum(1) maximum(100)
// @Success 200 {object} models.ListAgentsResponse "List of agents"
// @Failure 400 {object} models.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents [get]
func (c *AgentController) ListAgents(ctx *gin.Context) {
	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.ListAgentsRequest](ctx)
	if !exists {
		// Fall back to manual binding for backward compatibility
		var requestData models.ListAgentsRequest
		if err := ctx.ShouldBindQuery(&requestData); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Details: err.Error(),
			}
			ctx.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		request = requestData
	}

	response, err := c.agentService.ListAgents(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list agents",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateAgent handles PUT /agents/:agent_id
// @Summary Update an existing agent
// @Description Update an agent's configuration and settings
// @Tags agents
// @Accept json
// @Produce json
// @Param agent_id path string true "Agent ID"
// @Param request body models.UpdateAgentRequest true "Agent update request"
// @Success 200 {object} models.UpdateAgentResponse "Agent updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 404 {object} models.ErrorResponse "Agent not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /agents/{agent_id} [put]
func (c *AgentController) UpdateAgent(ctx *gin.Context) {
	agentID := ctx.Param("agent_id")
	if agentID == "" {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Agent ID is required",
			Details: "Agent ID must be provided in the URL path",
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Try to get the validated request from context first (new pattern)
	request, exists := middleware.GetValidatedRequest[models.UpdateAgentRequest](ctx)
	if !exists {
		// Fall back to manual binding for backward compatibility
		var requestData models.UpdateAgentRequest
		if err := ctx.ShouldBindJSON(&requestData); err != nil {
			errorResponse := models.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Invalid request body",
				Details: err.Error(),
			}
			ctx.JSON(http.StatusBadRequest, errorResponse)
			return
		}
		request = requestData
	}

	// Set the agent ID from the URL path
	request.AgentID = agentID

	response, err := c.agentService.UpdateAgent(ctx.Request.Context(), request)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update agent",
			Details: err.Error(),
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
