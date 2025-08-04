package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
)

// SetupRoutesWithAuth configures all routes with authentication middleware
func SetupRoutesWithAuth(
	router *gin.Engine,
	authProvider auth.AuthProvider,
	projectController *controllers.ProjectController,
	codebaseController *controllers.CodebaseController,
	agentController *controllers.AgentController,
	healthController *controllers.HealthController,
) {
	// Set up authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(authProvider)

	// Apply auth middleware globally (except for public endpoints)
	router.Use(authMiddleware.Handle())

	// Set up all routes
	SetupProjectRoutes(router, projectController)
	SetupCodebaseRoutes(router, codebaseController)
	SetupAgentRoutes(router, agentController)
	SetupHealthRoutes(router, healthController)
}

// SetupProjectRoutes configures the project routes with generic validation middleware
// This demonstrates the new annotation-based validation approach
func SetupProjectRoutes(router *gin.Engine, controller *controllers.ProjectController) {
	projectGroup := router.Group("/api/v1/projects")
	{
		// CREATE - validate JSON body using struct tags
		// The middleware automatically validates based on the struct tags in CreateProjectRequest
		projectGroup.POST("",
			middleware.NewJSONValidationMiddleware[models.CreateProjectRequest]().Handle(),
			controller.CreateProject,
		)

		// LIST - validate query parameters using struct tags
		// The middleware automatically validates based on the struct tags in ListProjectsRequest
		projectGroup.GET("",
			middleware.NewQueryValidationMiddleware[models.ListProjectsRequest]().Handle(),
			controller.ListProjects,
		)

		// GET by ID - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in GetProjectRequest
		projectGroup.GET("/:project_id",
			middleware.NewURIValidationMiddleware[models.GetProjectRequest]().Handle(),
			controller.GetProject,
		)

		// UPDATE - validate both URI and JSON using struct tags
		// The middleware automatically validates based on the struct tags in UpdateProjectRequest
		projectGroup.PUT("/:project_id",
			middleware.NewCombinedValidationMiddleware[models.UpdateProjectRequest]().Handle(),
			controller.UpdateProject,
		)

		// DELETE - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in DeleteProjectRequest
		projectGroup.DELETE("/:project_id",
			middleware.NewURIValidationMiddleware[models.DeleteProjectRequest]().Handle(),
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
			middleware.NewCombinedValidationMiddleware[models.CreateCodebaseRequest]().Handle(),
			controller.CreateCodebase,
		)
	}

	// Direct codebase routes
	codebaseGroup := router.Group("/api/v1/codebases")
	{
		// LIST - validate query parameters using struct tags
		codebaseGroup.GET("",
			middleware.NewQueryValidationMiddleware[models.ListCodebasesRequest]().Handle(),
			controller.ListCodebases,
		)

		// GET by ID - validate URI parameters using struct tags
		codebaseGroup.GET("/:id",
			middleware.NewURIValidationMiddleware[models.GetCodebaseRequest]().Handle(),
			controller.GetCodebase,
		)

		// UPDATE - validate both URI and JSON using struct tags
		codebaseGroup.PUT("/:id",
			middleware.NewCombinedValidationMiddleware[models.UpdateCodebaseRequest]().Handle(),
			controller.UpdateCodebase,
		)

		// DELETE - validate URI parameters using struct tags
		codebaseGroup.DELETE("/:id",
			middleware.NewURIValidationMiddleware[models.DeleteCodebaseRequest]().Handle(),
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
			middleware.NewJSONValidationMiddleware[models.CreateAgentRequest]().Handle(),
			controller.CreateAgent,
		)

		// LIST - validate query parameters using struct tags
		agentGroup.GET("",
			middleware.NewQueryValidationMiddleware[models.ListAgentsRequest]().Handle(),
			controller.ListAgents,
		)

		// GET by ID - validate URI parameters using struct tags
		agentGroup.GET("/:id",
			middleware.NewURIValidationMiddleware[models.GetAgentRequest]().Handle(),
			controller.GetAgent,
		)

		// DELETE - validate URI parameters using struct tags
		agentGroup.DELETE("/:id",
			middleware.NewURIValidationMiddleware[models.DeleteAgentRequest]().Handle(),
			controller.DeleteAgent,
		)
	}
}

// SetupHealthRoutes configures the health routes with generic validation middleware
func SetupHealthRoutes(router *gin.Engine, controller *controllers.HealthController) {
	// Health check endpoint - no validation needed but follows the pattern
	router.GET("/health",
		middleware.NewQueryValidationMiddleware[models.HealthCheckRequest]().Handle(),
		controller.HealthCheck,
	)
}
