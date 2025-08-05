package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
)

// SetupAgentRoutes configures the agent routes with validation middleware
func SetupAgentRoutes(router *gin.Engine, controller *controllers.AgentController) {
	agentGroup := router.Group("/api/v1/agents")
	{
		// CREATE - create a new agent
		agentGroup.POST("", controller.CreateAgent)

		// LIST - list agents for a project
		agentGroup.GET("", controller.ListAgents)

		// GET by ID - get agent details
		agentGroup.GET("/:agent_id", controller.GetAgent)

		// UPDATE - update an existing agent
		agentGroup.PUT("/:agent_id", controller.UpdateAgent)

		// DELETE - delete an agent
		agentGroup.DELETE("/:agent_id", controller.DeleteAgent)
	}
}
