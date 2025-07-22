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
	"github.com/kazemisoroush/code-refactoring-tool/api/services"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/logging"
	pkgRepo "github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
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

	// Initialize repositories and services
	agentRepository := repository.NewDynamoDBAgentRepository(
		cfg.AWSConfig,
		"", // Use default table name
	)

	// Initialize git repository (this will be used as a template)
	gitRepo := pkgRepo.NewGitHubRepo(cfg.Git)

	// Initialize S3 dataStore
	dataStore := storage.NewS3DataStore(cfg.AWSConfig, cfg.S3BucketName, gitRepo.GetPath())

	// Initialize storage
	storageService := storage.NewRDSPostgresStorage(cfg.AWSConfig, cfg.RDSPostgres.SchemaEnsureLambdaARN)

	// Initialize RAG pipeline
	ragService := rag.NewBedrockRAG(cfg.AWSConfig, gitRepo.GetPath(), cfg.KnowledgeBaseServiceRoleARN, cfg.RDSPostgres)

	// Initialize RAG builder
	ragBuilder := builder.NewBedrockRAGBuilder(
		gitRepo.GetPath(),
		dataStore,
		storageService,
		ragService,
	)

	// Initialize agent builder
	agentBuilder := builder.NewBedrockAgentBuilder(
		cfg.AWSConfig,
		gitRepo.GetPath(),
		cfg.AgentServiceRoleARN,
	)

	// Initialize service layer
	agentService := services.NewAgentService(cfg.Git, ragBuilder, agentBuilder, gitRepo, agentRepository)

	// Initialize controllers
	agentController := controllers.NewAgentController(agentService)

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(middleware.CognitoConfig{
		UserPoolID: cfg.Cognito.UserPoolID,
		Region:     cfg.Cognito.Region,
		ClientID:   cfg.Cognito.ClientID,
	})

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(authMiddleware.RequireAuth())

	// Setup API routes
	v1 := router.Group("/api/v1")
	{
		// Agent routes
		v1.POST("/agent/create", agentController.CreateAgent)
		v1.GET("/agent/:id", agentController.GetAgent)
		v1.DELETE("/agent/:id", agentController.DeleteAgent)
		v1.GET("/agents", agentController.ListAgents)
	}

	// Setup Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "code-refactor-tool-api",
			"version": "1.0.0",
		})
	})

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
