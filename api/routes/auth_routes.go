package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, authController *controllers.AuthController, authMiddleware middleware.Middleware) {
	authGroup := router.Group("/auth")
	{
		// Public routes (no authentication required)
		authGroup.POST("/signup", authController.SignUp)
		authGroup.POST("/signin", authController.SignIn)
		authGroup.POST("/refresh", authController.RefreshToken)
		authGroup.POST("/signout", authController.SignOut)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(authMiddleware.Handle())
		{
			protected.GET("/users", authController.ListUsers)
		}
	}
}
