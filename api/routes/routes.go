package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// SetupProjectRoutes configures the project routes with generic validation middleware
// This demonstrates the new annotation-based validation approach
func SetupProjectRoutes(router *gin.Engine, controller *controllers.ProjectController) {
	projectGroup := router.Group("/api/v1/projects")
	{
		// CREATE - validate JSON body using struct tags
		// The middleware automatically validates based on the struct tags in CreateProjectRequest
		projectGroup.POST("",
			middleware.ValidateJSON[models.CreateProjectRequest](),
			controller.CreateProject,
		)

		// LIST - validate query parameters using struct tags
		// The middleware automatically validates based on the struct tags in ListProjectsRequest
		projectGroup.GET("",
			middleware.ValidateQuery[models.ListProjectsRequest](),
			controller.ListProjects,
		)

		// GET by ID - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in GetProjectRequest
		projectGroup.GET("/:id",
			middleware.ValidateURI[models.GetProjectRequest](),
			controller.GetProject,
		)

		// UPDATE - validate both URI and JSON using struct tags
		// The middleware automatically validates based on the struct tags in UpdateProjectRequest
		projectGroup.PUT("/:id",
			middleware.ValidateCombined[models.UpdateProjectRequest](),
			controller.UpdateProject,
		)

		// DELETE - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in DeleteProjectRequest
		projectGroup.DELETE("/:id",
			middleware.ValidateURI[models.DeleteProjectRequest](),
			controller.DeleteProject,
		)
	}
}

// Example of how the same validation middleware can be extended for other entities
// Just create your models with appropriate validation tags and use the generic middleware
/*
func SetupAgentRoutes(router *gin.Engine, controller *controllers.AgentController) {
	agentGroup := router.Group("/api/v1/agents")
	{
		// CREATE - validate JSON body using struct tags
		agentGroup.POST("",
			middleware.ValidateJSON[models.CreateAgentRequest](),
			controller.CreateAgent,
		)

		// LIST - validate query parameters using struct tags
		agentGroup.GET("",
			middleware.ValidateQuery[models.ListAgentsRequest](),
			controller.ListAgents,
		)

		// GET by ID - validate URI parameters using struct tags
		agentGroup.GET("/:id",
			middleware.ValidateURI[models.GetAgentRequest](),
			controller.GetAgent,
		)
	}
}
*/
