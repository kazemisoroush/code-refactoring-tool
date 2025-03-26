package main

import (
	"context"
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/agent"
	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/kazemisoroush/code-refactor-tool/pkg/planner"
	"github.com/kazemisoroush/code-refactor-tool/pkg/repository"
	"github.com/kazemisoroush/code-refactor-tool/pkg/workflow"
)

// main is the entry point for the application.
func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a, err := analyzer.NewGolangCIAnalyzer()
	if err != nil {
		log.Fatalf("failed to create analyzer: %v", err)
	}

	r := repository.NewGitHubRepo(cfg.RepoURL, cfg.GitToken)
	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}

	agnt := agent.NewAWSBedrockAgent(cfg.AWSConfig, "TODO:modelID")

	p := planner.NewAIPlanner(agnt)

	wf, err := workflow.NewWorkflow(cfg, a, r, p)
	if err != nil {
		log.Fatalf("failed to create workflow: %v", err)
	}

	err = wf.Run(ctx)
	if err != nil {
		log.Fatalf("workflow failed: %v", err)
	}
}
