package models

import (
	"strconv"
	"strings"
)

type GolangCIIssue struct {
	FromLinter           string           `json:"FromLinter"`
	Text                 string           `json:"Text"`
	Severity             string           `json:"Severity"`
	SourceLines          []string         `json:"SourceLines"`
	Pos                  GolangCIPosition `json:"Pos"`
	ExpectNoLint         bool             `json:"ExpectNoLint"`
	ExpectedNoLintLinter string           `json:"ExpectedNoLintLinter"`
}

type GolangCIPosition struct {
	Filename string `json:"Filename"`
	Offset   int    `json:"Offset"`
	Line     int    `json:"Line"`
	Column   int    `json:"Column"`
}

type GolangCILinter struct {
	Name             string `json:"Name"`
	Enabled          bool   `json:"Enabled,omitempty"`
	EnabledByDefault bool   `json:"EnabledByDefault,omitempty"`
}

type GolangCIReport struct {
	Linters []GolangCILinter `json:"Linters"`
}

type GolangCILintReport struct {
	Issues []GolangCIIssue `json:"Issues"`
	Report GolangCIReport  `json:"Report"`
}

func (g *GolangCILintReport) GetCyclomaticComplexity() int {
	count := 0
	for _, issue := range g.Issues {
		if issue.FromLinter == "gocyclo" {
			count++
		}
	}
	return count
}

func (g *GolangCILintReport) GetDuplicateCode() int {
	count := 0
	for _, issue := range g.Issues {
		if issue.FromLinter == "dupl" {
			count++
		}
	}
	return count
}

func (g *GolangCILintReport) GetTestCoverage() float64 {
	for _, issue := range g.Issues {
		if strings.Contains(issue.Text, "coverage:") {
			parts := strings.Fields(issue.Text)
			for _, part := range parts {
				if strings.HasSuffix(part, "%") {
					val, err := strconv.ParseFloat(strings.TrimSuffix(part, "%"), 64)
					if err == nil {
						return val
					}
				}
			}
		}
	}
	return 0.0
}

func (g *GolangCILintReport) GetFunctionCount() int {
	count := 0
	for _, issue := range g.Issues {
		if issue.FromLinter == "funlen" {
			count++
		}
	}
	return count
}

func (g *GolangCILintReport) GetLongFunctions() int {
	count := 0
	for _, issue := range g.Issues {
		if issue.FromLinter == "gocognit" {
			count++
		}
	}
	return count
}

func (g *GolangCILintReport) GetDeadCodeCount() int {
	count := 0
	for _, issue := range g.Issues {
		if issue.FromLinter == "deadcode" || issue.FromLinter == "unused" {
			count++
		}
	}
	return count
}
