package models_test

import (
	"testing"

	"github.com/kazemisoroush/code-refactor-tool/pkg/analyzer/models"
)

// TestGetCyclomaticComplexity
func TestGetCyclomaticComplexityExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{FromLinter: "gocyclo"},
			{FromLinter: "gocyclo"},
		},
	}

	expected := 2
	result := report.GetCyclomaticComplexity()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestGetCyclomaticComplexityNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0
	result := report.GetCyclomaticComplexity()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

// TestGetDuplicateCode
func TestGetDuplicateCodeExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{FromLinter: "dupl"},
			{FromLinter: "dupl"},
		},
	}

	expected := 2
	result := report.GetDuplicateCode()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestGetDuplicateCodeNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0
	result := report.GetDuplicateCode()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

// TestGetTestCoverage
func TestGetTestCoverageExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{Text: "coverage: 85.4%"},
		},
	}

	expected := 85.4
	result := report.GetTestCoverage()
	if result != expected {
		t.Errorf("Expected %.1f, got %.1f", expected, result)
	}
}

func TestGetTestCoverageNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0.0
	result := report.GetTestCoverage()
	if result != expected {
		t.Errorf("Expected %.1f, got %.1f", expected, result)
	}
}

// TestGetFunctionCount
func TestGetFunctionCountExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{FromLinter: "funlen"},
			{FromLinter: "funlen"},
		},
	}

	expected := 2
	result := report.GetFunctionCount()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestGetFunctionCountNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0
	result := report.GetFunctionCount()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

// TestGetLongFunctions
func TestGetLongFunctionsExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{FromLinter: "gocognit"},
			{FromLinter: "gocognit"},
		},
	}

	expected := 2
	result := report.GetLongFunctions()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestGetLongFunctionsNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0
	result := report.GetLongFunctions()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

// TestGetDeadCodeCount
func TestGetDeadCodeCountExists(t *testing.T) {
	report := models.GolangCILintReport{
		Issues: []models.GolangCIIssue{
			{FromLinter: "deadcode"},
			{FromLinter: "unused"},
		},
	}

	expected := 2
	result := report.GetDeadCodeCount()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}

func TestGetDeadCodeCountNotExists(t *testing.T) {
	report := models.GolangCILintReport{}

	expected := 0
	result := report.GetDeadCodeCount()
	if result != expected {
		t.Errorf("Expected %d, got %d", expected, result)
	}
}
