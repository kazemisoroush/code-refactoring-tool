// Package repository provides a generic interface for interacting with code repositories.
package repository

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kazemisoroush/code-refactoring-tool/pkg/config"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	RepoURL string
	Token   string
	Author  string
	Email   string
	repo    *git.Repository
	path    string
}

// NewGitHubRepo creates a new GitHub repository instance
func NewGitHubRepo(git config.GitConfig) Repository {
	repoName := path.Base(strings.TrimSuffix(git.RepoURL, ".git"))
	return &GitHubRepo{
		RepoURL: git.RepoURL,
		Token:   git.Token,
		Author:  git.Author,
		Email:   git.Email,
		path:    repoName,
	}
}

// GetPath implements Repository.
func (g *GitHubRepo) GetPath() string {
	return g.path
}

// Clone clones the repository to the local filesystem
func (g *GitHubRepo) Clone() error {
	repo, err := git.PlainClone(g.path, false, &git.CloneOptions{
		URL:      g.RepoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	g.repo = repo
	return nil
}

// CheckoutBranch creates a new branch and checks it out
func (g *GitHubRepo) CheckoutBranch(branchName string) error {
	wt, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	branchRef := plumbing.NewBranchReferenceName(branchName)
	return wt.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Create: true,
	})
}

// Commit stages and commits all changes with the provided message
func (g *GitHubRepo) Commit(message string) error {
	wt, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	if _, err := wt.Add("."); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.Author,
			Email: g.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

// Push pushes commits to the remote repository
func (g *GitHubRepo) Push() error {
	return g.repo.Push(&git.PushOptions{
		RemoteURL:  g.RepoURL,
		RemoteName: "origin",
		Force:      true,
		Auth: &gitHttp.BasicAuth{
			Username: g.Author,
			Password: g.Token,
		},
	})
}

// CreatePR creates a new pull request if one does not already exist.
func (g *GitHubRepo) CreatePR(title, description, sourceBranch, targetBranch string) (string, error) {
	repoParts := strings.Split(strings.TrimPrefix(g.RepoURL, "https://github.com/"), "/")
	if len(repoParts) != 2 {
		return "", fmt.Errorf("invalid repo URL format: %s", g.RepoURL)
	}
	owner := repoParts[0]
	repo := strings.TrimSuffix(repoParts[1], ".git")

	// Step 1: Check if PR already exists
	checkURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?head=%s:%s&base=%s", owner, repo, owner, sourceBranch, targetBranch)
	checkCmd := exec.Command("curl", "-s", "-H", fmt.Sprintf("Authorization: token %s", g.Token), checkURL)
	existingPRBytes, err := checkCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to check for existing PR: %w", err)
	}
	if strings.Contains(string(existingPRBytes), "\"url\"") {
		fmt.Println("ℹ️  Pull request already exists. Skipping creation.")
		return "PR already exists", nil
	}

	// Step 2: Create new PR
	createURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	payload := fmt.Sprintf(`{
		"title": "%s",
		"body": "%s",
		"head": "%s",
		"base": "%s"
	}`, title, description, sourceBranch, targetBranch)

	createCmd := exec.Command("curl", "-s", "-X", "POST",
		"-H", fmt.Sprintf("Authorization: token %s", g.Token),
		"-H", "Accept: application/vnd.github.v3+json",
		createURL, "-d", payload)

	output, err := createCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}
	return string(output), nil
}

// Cleanup deletes the repository from the filesystem.
func (g *GitHubRepo) Cleanup() error {
	err := os.RemoveAll(g.path)
	if err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}
	return nil
}
