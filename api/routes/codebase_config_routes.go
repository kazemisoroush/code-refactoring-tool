package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// SetupCodebaseConfigRoutes configures the codebase configuration routes with generic validation middleware
func SetupCodebaseConfigRoutes(router *gin.Engine, controller *controllers.CodebaseConfigController) {
	codebaseConfigGroup := router.Group("/api/v1/codebase-configs")
	{
		// CREATE - validate JSON body using struct tags
		// The middleware automatically validates based on the struct tags in CreateCodebaseConfigRequest
		codebaseConfigGroup.POST("",
			middleware.NewJSONValidationMiddleware[models.CreateCodebaseConfigRequest]().Handle(),
			controller.CreateCodebaseConfig,
		)

		// LIST - validate query parameters using struct tags
		// The middleware automatically validates based on the struct tags in ListCodebaseConfigsRequest
		codebaseConfigGroup.GET("",
			middleware.NewQueryValidationMiddleware[models.ListCodebaseConfigsRequest]().Handle(),
			controller.ListCodebaseConfigs,
		)

		// GET by ID - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in GetCodebaseConfigRequest
		codebaseConfigGroup.GET("/:config_id",
			middleware.NewURIValidationMiddleware[models.GetCodebaseConfigRequest]().Handle(),
			controller.GetCodebaseConfig,
		)

		// UPDATE - validate both URI and JSON using struct tags
		// The middleware automatically validates based on the struct tags in UpdateCodebaseConfigRequest
		codebaseConfigGroup.PUT("/:config_id",
			middleware.NewCombinedValidationMiddleware[models.UpdateCodebaseConfigRequest]().Handle(),
			controller.UpdateCodebaseConfig,
		)

		// DELETE - validate URI parameters using struct tags
		// The middleware automatically validates based on the struct tags in DeleteCodebaseConfigRequest
		codebaseConfigGroup.DELETE("/:config_id",
			middleware.NewURIValidationMiddleware[models.DeleteCodebaseConfigRequest]().Handle(),
			controller.DeleteCodebaseConfig,
		)
	}
}
