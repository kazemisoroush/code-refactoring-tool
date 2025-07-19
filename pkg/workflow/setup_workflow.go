// Package workflow provides the main workflow for the code analysis tool.
package workflow

import (
	"context"
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactoring-tool/pkg/ai/builder"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/repository"
)

// SetupWorkflow represents a workflow for setting up AI resources
type SetupWorkflow struct {
	Config       config.Config
	Repository   repository.Repository
	RAGBuilder   builder.RAGBuilder
	AgentBuilder builder.AgentBuilder

	// Resource IDs created during setup
	VectorStoreID string
	RAGID         string
	AgentID       string
	AgentVersion  string
}

// NewSetupWorkflow creates a new SetupWorkflow instance
func NewSetupWorkflow(
	cfg config.Config,
	repo repository.Repository,
	ragBuilder builder.RAGBuilder,
	agentBuilder builder.AgentBuilder,
) (Workflow, error) {
	return &SetupWorkflow{
		Config:       cfg,
		Repository:   repo,
		RAGBuilder:   ragBuilder,
		AgentBuilder: agentBuilder,
	}, nil
}

// Run executes the setup workflow to provision AI resources
func (s *SetupWorkflow) Run(ctx context.Context) error {
	log.Println("Running setup workflow...")

	// 1. Clone the repository
	log.Println("Cloning repository...")
	err := s.Repository.Clone(ctx)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	log.Printf("✅ Repository cloned successfully")

	// 2. Build the RAG pipeline
	log.Println("Building RAG pipeline...")
	ragID, err := s.RAGBuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build RAG pipeline: %w", err)
	}
	s.RAGID = ragID
	s.VectorStoreID = ragID // In Bedrock, the KB ID serves as both RAG ID and vector store ID
	log.Printf("✅ RAG pipeline built successfully. RAG ID: %s", ragID)

	// 3. Build the agent
	log.Printf("Building agent with RAG ID: %s", ragID)
	agentID, agentVersion, err := s.AgentBuilder.Build(ctx, ragID)
	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}
	s.AgentID = agentID
	s.AgentVersion = agentVersion
	log.Printf("✅ Agent built successfully. Agent ID: %s, Version: %s", agentID, agentVersion)

	log.Println("✅ Setup workflow completed successfully")
	return nil
}

// GetResourceIDs returns the resource IDs created during setup
func (s *SetupWorkflow) GetResourceIDs() (vectorStoreID, ragID, agentID, agentVersion string) {
	return s.VectorStoreID, s.RAGID, s.AgentID, s.AgentVersion
}
