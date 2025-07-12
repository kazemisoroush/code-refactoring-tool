// Package main provides the entry point for the application.
package main

import (
	"context"
	"log"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/rag"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/storage"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

// main is the entry point for the application.
func main() {
	// Load environment + AWS config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Initialize code repository
	repo := repository.NewGitHubRepo(cfg.Git)

	// Initialize S3 dataStore
	dataStore := storage.NewS3Storage(cfg.AWSConfig, cfg.S3BucketName, repo.GetPath())

	// Initialize RAG pipeline
	rag := rag.NewBedrockRAG(cfg.AWSConfig, repo.GetPath(), cfg.KnowledgeBaseServiceRoleARN, cfg.RDSPostgres)

	// Initialize RAG builder with AWS configuration
	ragBuilder := builder.NewBedrockRAGBuilder(
		repo.GetPath(),
		dataStore,
		rag,
	)

	// Initialize agent builder with AWS configuration
	agentBuilder := builder.NewBedrockAgentBuilder(
		cfg.AWSConfig,
		repo.GetPath(),
		cfg.AgentServiceRoleARN,
	)

	// Compose the full workflow
	wf, err := workflow.NewFixWorkflow(cfg, repo, ragBuilder, agentBuilder)
	if err != nil {
		log.Fatalf("❌ failed to create workflow: %v", err)
	}

	// Run the workflow
	if err := wf.Run(ctx); err != nil {
		log.Fatalf("❌ workflow failed: %v", err)
	}

	log.Println("✅ Workflow completed successfully.")
}
