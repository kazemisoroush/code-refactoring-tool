// Package main provides the entry point for the application.
package main

import (
	"context"
	"log"
	"time"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/agent"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/patcher"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/workflow"
)

const (
	claude3Sonnet  = "anthropic.claude-3-sonnet-20240229-v1:0"
	claude35Sonnet = "anthropic.claude-3-5-sonnet-20240620-v1:0"
	mistral7B      = "mistral.mistral-7b-instruct-v0:2"
	mistralLarge   = "mistral.mistral-large-2402-v1:0"
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

	// Initialize static analyzer
	anlzr, err := analyzer.NewGolangCIAnalyzer()
	if err != nil {
		log.Fatalf("❌ failed to create analyzer: %v", err)
	}

	// Initialize code repository
	repo := repository.NewGitHubRepo(cfg.Git)

	// Use a valid Bedrock model ID (Claude 3 Sonnet here as an example)
	agnt := agent.NewAWSBedrockAgent(cfg.AWSConfig, claude35Sonnet)

	// Create planner using AI agent
	plnr := planner.NewAIPlanner(agnt)

	ptchr := patcher.NewFilePatcher()

	// Compose the full workflow
	wf, err := workflow.NewWorkflow(cfg, anlzr, repo, plnr, ptchr)
	if err != nil {
		log.Fatalf("❌ failed to create workflow: %v", err)
	}

	// Run the workflow
	if err := wf.Run(ctx); err != nil {
		log.Fatalf("❌ workflow failed: %v", err)
	}

	log.Println("✅ Workflow completed successfully.")
}
