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

// SetupCodebaseRoutes configures the codebase routes with generic validation middleware
func SetupCodebaseRoutes(router *gin.Engine, controller *controllers.CodebaseController) {
	// Codebase routes nested under projects
	projectCodebaseGroup := router.Group("/api/v1/projects/:project_id/codebases")
	{
		// CREATE - validate combined URI (project_id) and JSON body using struct tags
		projectCodebaseGroup.POST("",
			middleware.ValidateCombined[models.CreateCodebaseRequest](),
			controller.CreateCodebase,
		)
	}

	// Direct codebase routes
	codebaseGroup := router.Group("/api/v1/codebases")
	{
		// LIST - validate query parameters using struct tags
		codebaseGroup.GET("",
			middleware.ValidateQuery[models.ListCodebasesRequest](),
			controller.ListCodebases,
		)

		// GET by ID - validate URI parameters using struct tags
		codebaseGroup.GET("/:id",
			middleware.ValidateURI[models.GetCodebaseRequest](),
			controller.GetCodebase,
		)

		// UPDATE - validate both URI and JSON using struct tags
		codebaseGroup.PUT("/:id",
			middleware.ValidateCombined[models.UpdateCodebaseRequest](),
			controller.UpdateCodebase,
		)

		// DELETE - validate URI parameters using struct tags
		codebaseGroup.DELETE("/:id",
			middleware.ValidateURI[models.DeleteCodebaseRequest](),
			controller.DeleteCodebase,
		)
	}
}

// SetupAgentRoutes configures the agent routes with generic validation middleware
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

		// DELETE - validate URI parameters using struct tags
		agentGroup.DELETE("/:id",
			middleware.ValidateURI[models.DeleteAgentRequest](),
			controller.DeleteAgent,
		)
	}
}

// SetupHealthRoutes configures the health routes with generic validation middleware
func SetupHealthRoutes(router *gin.Engine, controller *controllers.HealthController) {
	// Health check endpoint - no validation needed but follows the pattern
	router.GET("/health",
		middleware.ValidateQuery[models.HealthCheckRequest](),
		controller.HealthCheck,
	)
}
