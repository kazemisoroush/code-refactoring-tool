package models

// GoBuildIssues Issues from go build CLI.
type GoBuildIssues struct {
	ImportPath string `json:"ImportPath"`
	Action     string `json:"Action"`
	Output     string `json:"Output"`
}

// GoTestIssue go test issue.
type GoTestIssue struct {
	Time        string `json:"Time"`
	Action      string `json:"Action"`
	Package     string `json:"Package"`
	Elapsed     int    `json:"Elapsed,omitempty"`
	FailedBuild string `json:"FailedBuild,omitempty"`
	Output      string `json:"Output,omitempty"`
}
