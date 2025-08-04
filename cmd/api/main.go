// Package main provides the entry point for the API server
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/api/routes"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/auth"
	appconfig "github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/factory"
)

// @title Code Refactor Tool API
// @version 1.0
// @description API for creating and managing AI-powered code refactoring agents
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load configuration
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Give outstanding requests a deadline for completion
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer shutdownCancel()

	// Initialize Postgres config for repositories
	postgresConfig := repository.PostgresConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		Database: cfg.Postgres.Database,
		Username: cfg.Postgres.Username,
		Password: cfg.Postgres.Password,
		SSLMode:  cfg.Postgres.SSLMode,
	}

	// Initialize agent repository
	agentRepository, err := repository.NewPostgresAgentRepository(postgresConfig, appconfig.DefaultAgentsTableName)
	if err != nil {
		slog.Error("failed to initialize agent repository", "error", err)
	}

	// Initialize project repository
	projectRepository, err := repository.NewPostgresProjectRepository(postgresConfig, appconfig.DefaultProjectsTableName)
	if err != nil {
		slog.Error("failed to initialize project repository", "error", err)
		os.Exit(1)
	}

	// Initialize codebase repository
	codebaseRepository, err := repository.NewPostgresCodebaseRepository(postgresConfig, appconfig.DefaultCodebasesTableName)
	if err != nil {
		slog.Error("failed to initialize codebase repository", "error", err)
		os.Exit(1)
	}

	// Initialize task repository
	taskRepository, err := repository.NewPostgresTaskRepository(postgresConfig, appconfig.DefaultTasksTableName)
	if err != nil {
		slog.Error("failed to initialize task repository", "error", err)
		os.Exit(1)
	}

	// Initialize user repository
	userRepository, err := repository.NewPostgresUserRepository(postgresConfig, appconfig.DefaultUsersTableName)
	if err != nil {
		slog.Error("failed to initialize user repository", "error", err)
		os.Exit(1)
	}

	// TODO: Git config needs to come from API or Database
	aiFactory := factory.NewTaskExecutionFactory(cfg.AWSConfig, cfg.Git)

	// Initialize services with full dependency injection
	projectService := services.NewDefaultProjectService(projectRepository)
	codebaseService := services.NewDefaultCodebaseService(codebaseRepository)
	healthService := services.NewDefaultHealthService("code-refactor-tool-api", "1.0.0")

	// Agent repository temporarily nil for API-first development
	taskService := services.NewTaskService(
		taskRepository,
		projectRepository,
		agentRepository,
		codebaseRepository,
		aiFactory,
	)

	projectController := controllers.NewProjectController(projectService)
	codebaseController := controllers.NewCodebaseController(codebaseService)
	taskController := controllers.NewTaskController(taskService)
	healthController := controllers.NewHealthController(healthService)

	// Initialize AWS config
	awsConfig, err := config.LoadDefaultConfig(shutdownCtx, config.WithRegion(cfg.Cognito.Region))
	if err != nil {
		slog.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Initialize Cognito provider and authentication middleware
	cognitoProvider := auth.NewCognitoProvider(awsConfig, cfg.Cognito)

	// Initialize auth service with user repository and auth provider
	authService := services.NewAuthService(cognitoProvider, userRepository)

	// Initialize auth controller
	authController := controllers.NewAuthController(authService)

	authMiddleware := middleware.NewAuthMiddleware(cognitoProvider)

	// Initialize metrics middleware
	metricsMiddleware, err := middleware.NewMetricsMiddleware(appconfig.MetricsConfig{
		Namespace:   cfg.Metrics.Namespace,
		Region:      cfg.Metrics.Region,
		ServiceName: cfg.Metrics.ServiceName,
		Enabled:     cfg.Metrics.Enabled,
	})
	if err != nil {
		slog.Error("failed to initialize metrics middleware", "error", err)
		os.Exit(1)
	}

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add metrics middleware (before auth to capture all requests)
	router.Use(metricsMiddleware.Handle())

	// Add authentication middleware
	router.Use(authMiddleware.Handle())

	// Setup project routes with validation middleware
	routes.SetupProjectRoutes(router, projectController)

	// Setup codebase routes with validation middleware
	routes.SetupCodebaseRoutes(router, codebaseController)

	// Setup task routes with validation middleware - NEW!
	routes.SetupTaskRoutes(router, taskController)

	// Setup auth routes with authentication middleware
	routes.RegisterAuthRoutes(router, authController, authMiddleware)

	// Setup health routes with validation middleware
	routes.SetupHealthRoutes(router, healthController)

	// Setup Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}
