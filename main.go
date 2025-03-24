package main

import (
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/kazemisoroush/code-refactor-tool/pkg/repository"
	"github.com/kazemisoroush/code-refactor-tool/pkg/workflow"
)

// main is the entry point for the application.
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	a, err := analyzer.NewGoAnalyzer()
	if err != nil {
		log.Fatalf("failed to create analyzer: %v", err)
	}

	r := repository.NewGitHubRepo(cfg.RepoURL, cfg.GitToken)
	if err != nil {
		log.Fatalf("failed to create repository: %v", err)
	}

	wf, err := workflow.NewWorkflow(cfg, a, r)
	if err != nil {
		log.Fatalf("failed to create workflow: %v", err)
	}

	err = wf.Run()
	if err != nil {
		log.Fatalf("workflow failed: %v", err)
	}
}
