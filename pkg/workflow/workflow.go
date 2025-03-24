package workflow

import (
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer"
	"github.com/kazemisoroush/code-refactor-tool/pkg/config"
	"github.com/kazemisoroush/code-refactor-tool/pkg/repository"
)

// Workflow represents a code analysis workflow
type Workflow struct {
	Config     config.Config
	Analyzer   analyzer.CodeAnalyzer
	Repository repository.Repository
}

// NewWorkflow creates a new Workflow instance
func NewWorkflow(
	cfg config.Config,
	analyzer analyzer.CodeAnalyzer,
	repo repository.Repository,
) (*Workflow, error) {
	return &Workflow{
		Config:     cfg,
		Analyzer:   analyzer,
		Repository: repo,
	}, nil
}

// Run executes the code analysis workflow
func (o *Workflow) Run() error {
	log.Println("Running workflow...")

	// Clone the repository
	err := o.Repository.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Analyze the code
	analysisResult, err := o.Analyzer.AnalyzeCode(o.Repository.GetPath())
	if err != nil {
		return fmt.Errorf("failed to analyze code: %w", err)
	}

	// Extract metrics
	metrics, err := o.Analyzer.ExtractMetrics(analysisResult)
	if err != nil {
		return fmt.Errorf("failed to extract metrics: %w", err)
	}

	// Generate report
	report := o.Analyzer.GenerateReport(metrics)

	log.Printf("Workflow completed successfully! Report: %s", report.ToString())

	return nil
}
