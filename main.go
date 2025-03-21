package main

import (
	"fmt"
	"log"

	"github.com/kazemisoroush/code-refactor-tool/pkg/golang"
)

func main() {
	sourcePath := "./pkg/test/golang/bad_code_1.go"

	analyzer, err := golang.NewGoAnalyzer()
	if err != nil {
		log.Fatalf("❌ Error creating analyzer: %v", err)
	}

	fmt.Println("🔍 Running code analysis...")
	analysisResult, err := analyzer.AnalyzeCode(sourcePath)
	if err != nil {
		log.Fatalf("❌ Error running analysis: %v", err)
	}

	codeMetrics, err := analyzer.ExtractMetrics(analysisResult)
	if err != nil {
		log.Fatalf("❌ Error extracting metrics: %v", err)
	}

	report := analyzer.GenerateReport(codeMetrics)

	fmt.Println("\n📊 Analysis Report:")
	fmt.Printf("🔹 Language: %s\n", report.Language)
	fmt.Printf("🔹 Cyclomatic Complexity: %d\n", report.CodeMetrics.CyclomaticComplexity)
	fmt.Printf("🔹 Duplicate Code: %d\n", report.CodeMetrics.DuplicateCode)
	fmt.Printf("🔹 Test Coverage: %.2f%%\n", report.CodeMetrics.TestCoverage)
	fmt.Printf("🔹 Function Count: %d\n", report.CodeMetrics.FunctionCount)
	fmt.Printf("🔹 Long Functions: %d\n", report.CodeMetrics.LongFunctions)
	fmt.Printf("🔹 Dead Code Count: %d\n", report.CodeMetrics.DeadCodeCount)

	fmt.Println("\n🛠️ Refactoring Suggestions:")
	for _, suggestion := range report.Suggestions {
		fmt.Printf("  - %s\n", suggestion)
	}

	fmt.Println("\n✅ Done!")
}
