// Package repository provides a generic interface for interacting with code repositories.
package repository

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	RepoURL string
	Token   string
}

// NewGitHubRepo creates a new GitHub repository instance
func NewGitHubRepo(repoURL, token string) Repository {
	return &GitHubRepo{
		RepoURL: repoURL,
		Token:   token,
	}
}

// GetPath implements Repository.
func (g *GitHubRepo) GetPath() string {
	// Trim protocol (https:// or git@) and extract the last path component
	repoName := path.Base(strings.TrimSuffix(g.RepoURL, ".git"))
	return repoName
}

// Clone clones the repository to the local filesystem
func (g *GitHubRepo) Clone() error {
	cmd := exec.Command("git", "clone", g.RepoURL)
	return cmd.Run()
}

// CheckoutBranch creates a new branch and checks it out
func (g *GitHubRepo) CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	return cmd.Run()
}

// Commit Add adds all changes to the staging area
func (g *GitHubRepo) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-am", message)
	return cmd.Run()
}

// Push commits to the remote repository
func (g *GitHubRepo) Push() error {
	cmd := exec.Command("git", "push", "origin", "HEAD")
	return cmd.Run()
}

// CreatePR creates a new pull request
func (g *GitHubRepo) CreatePR(title, description, sourceBranch, targetBranch string) (string, error) {
	prURL := fmt.Sprintf("https://api.github.com/repos/%s/pulls", "OWNER/REPO")
	cmd := exec.Command("curl", "-X", "POST", "-H", fmt.Sprintf("Authorization: token %s", g.Token),
		"-H", "Accept: application/vnd.github.v3+json",
		prURL,
		"-d", fmt.Sprintf(`{"title":"%s", "body":"%s", "head":"%s", "base":"%s"}`, title, description, sourceBranch, targetBranch))

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Cleanup deletes the repository from the filesystem.
func (g *GitHubRepo) Cleanup() error {
	cmd := exec.Command("rm", "-rf", g.GetPath())
	return cmd.Run()
}
