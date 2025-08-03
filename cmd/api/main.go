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

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/kazemisoroush/code-refactoring-tool/api/controllers"
	"github.com/kazemisoroush/code-refactoring-tool/api/middleware"
	"github.com/kazemisoroush/code-refactoring-tool/api/repository"
	"github.com/kazemisoroush/code-refactoring-tool/api/routes"
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/factory"
	pkgRepo "github.com/kazemisoroush/code-refactoring-tool/pkg/codebase"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/logging"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup structured logging
	logging.SetupLogger(cfg.LogLevel)

	// Initialize Postgres config for repositories
	postgresConfig := repository.PostgresConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		Database: cfg.Postgres.Database,
		Username: cfg.Postgres.Username,
		Password: cfg.Postgres.Password,
		SSLMode:  cfg.Postgres.SSLMode,
	}

	// Initialize repositories and services
	agentRepository, err := repository.NewPostgresAgentRepository(postgresConfig, config.DefaultAgentsTableName)
	if err != nil {
		slog.Error("failed to initialize agent repository", "error", err)
		os.Exit(1)
	}

	// Initialize project repository
	projectRepository, err := repository.NewPostgresProjectRepository(postgresConfig, config.DefaultProjectsTableName)
	if err != nil {
		slog.Error("failed to initialize project repository", "error", err)
		os.Exit(1)
	}

	// Create projects table if it doesn't exist
	if projectRepo, ok := projectRepository.(*repository.PostgresProjectRepository); ok {
		if err := projectRepo.CreateTable(context.Background()); err != nil {
			slog.Error("failed to create projects table", "error", err)
			os.Exit(1)
		}
	}

	// Initialize codebase repository
	codebaseRepository, err := repository.NewPostgresCodebaseRepository(postgresConfig, config.DefaultCodebasesTableName)
	if err != nil {
		slog.Error("failed to initialize codebase repository", "error", err)
		os.Exit(1)
	}

	// Initialize git repository (this will be used as a template)
	gitRepo := pkgRepo.NewGitHubCodebase(cfg.Git)

	// Initialize AI factory (factory will create the appropriate services based on AI config)
	aiFactory := factory.NewAIProviderFactory(
		cfg.AWSConfig,
		&cfg.AI,
	)

	// Create builders using factory with no specific AI configuration (use platform defaults)
	ragBuilder, err := aiFactory.CreateRAGBuilder(nil, gitRepo.GetPath())
	if err != nil {
		slog.Error("failed to create RAG builder", "error", err)
		os.Exit(1)
	}

	agentBuilder, err := aiFactory.CreateAgentBuilder(nil, gitRepo.GetPath())
	if err != nil {
		slog.Error("failed to create agent builder", "error", err)
		os.Exit(1)
	}

	// Initialize service layer
	agentService := services.NewDefaultAgentService(cfg.Git, ragBuilder, agentBuilder, gitRepo, agentRepository)
	projectService := services.NewDefaultProjectService(projectRepository)
	codebaseService := services.NewDefaultCodebaseService(codebaseRepository)
	healthService := services.NewDefaultHealthService("code-refactor-tool-api", "1.0.0")

	// Initialize controllers
	agentController := controllers.NewAgentController(agentService)
	projectController := controllers.NewProjectController(projectService)
	codebaseController := controllers.NewCodebaseController(codebaseService)
	healthController := controllers.NewHealthController(healthService)

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(config.CognitoConfig{
		UserPoolID: cfg.Cognito.UserPoolID,
		Region:     cfg.Cognito.Region,
		ClientID:   cfg.Cognito.ClientID,
	})

	// Initialize metrics middleware
	metricsMiddleware, err := middleware.NewMetricsMiddleware(config.MetricsConfig{
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
	router.Use(metricsMiddleware.Layer())
	router.Use(metricsMiddleware.SetMetricsInContext())

	// Add authentication middleware
	router.Use(authMiddleware.Handle())

	// Setup API routes
	v1 := router.Group("/api/v1")
	{
		// Keep old routes for backward compatibility (these don't conflict with new ones)
		v1.POST("/agent/create", agentController.CreateAgent)
		v1.GET("/agent/:id", agentController.GetAgent)
		v1.DELETE("/agent/:id", agentController.DeleteAgent)
		// Remove this line as it conflicts with the new route
		// v1.GET("/agents", agentController.ListAgents)
	}

	// Setup agent routes with validation middleware (new standardized routes)
	routes.SetupAgentRoutes(router, agentController)

	// Setup project routes with validation middleware
	routes.SetupProjectRoutes(router, projectController)

	// Setup codebase routes with validation middleware
	routes.SetupCodebaseRoutes(router, codebaseController)

	// Setup health routes with validation middleware
	routes.SetupHealthRoutes(router, healthController)

	// Setup Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check (keep old route for backward compatibility)
	// router.GET("/health", healthController.HealthCheck)

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

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}
