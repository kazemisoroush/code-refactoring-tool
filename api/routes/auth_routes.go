package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/models"
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
		authGroup.POST("/confirm", middleware.NewJSONValidationMiddleware[models.ConfirmEmailRequest]().Handle(), authController.ConfirmEmail)
		authGroup.POST("/forgot-password", middleware.NewJSONValidationMiddleware[models.ForgotPasswordRequest]().Handle(), authController.ForgotPassword)
		authGroup.POST("/reset-password", middleware.NewJSONValidationMiddleware[models.ResetPasswordRequest]().Handle(), authController.ResetPassword)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(authMiddleware.Handle())
		{
			protected.GET("/me", authController.GetMe)
			protected.GET("/users", authController.ListUsers)
			protected.POST("/users", authController.CreateUser)
			protected.GET("/users/:id", authController.GetUser)
			protected.PUT("/users/:id", authController.UpdateUser)
			protected.DELETE("/users/:id", authController.DeleteUser)
		}
	}
}
