package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
)

// SetupTaskRoutes sets up the task-related routes
func SetupTaskRoutes(router *gin.Engine, taskController *controllers.TaskController) {
	// Task routes - project-scoped
	v1 := router.Group("/api/v1")
	{
		// Project-scoped task routes
		projects := v1.Group("/projects/:project_id")
		{
			// Create task for a project
			projects.POST("/tasks",
				middleware.NewJSONValidationMiddleware[models.CreateTaskRequest]().Handle(),
				taskController.CreateTask,
			)

			// List tasks for a project
			projects.GET("/tasks",
				middleware.NewQueryValidationMiddleware[models.ListTasksRequest]().Handle(),
				taskController.ListTasks,
			)

			// Execute task immediately for a project
			projects.POST("/tasks/execute",
				middleware.NewJSONValidationMiddleware[models.ExecuteTaskRequest]().Handle(),
				taskController.ExecuteTask,
			)
		}

		// Global task routes (not project-scoped)
		tasks := v1.Group("/tasks")
		{
			// Get specific task by ID
			tasks.GET("/:id",
				middleware.NewURIValidationMiddleware[models.GetTaskRequest]().Handle(),
				taskController.GetTask,
			)

			// Update task by ID
			tasks.PUT("/:id",
				middleware.NewJSONValidationMiddleware[models.UpdateTaskRequest]().Handle(),
				taskController.UpdateTask,
			)

			// Delete task by ID
			tasks.DELETE("/:id", taskController.DeleteTask)
		}
	}
}
