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

// CreatePR creates a new pull request (still using curl for simplicity)
func (g *GitHubRepo) CreatePR(title, description, sourceBranch, targetBranch string) (string, error) {
	prURL := fmt.Sprintf("https://api.github.com/repos/%s/pulls", strings.TrimSuffix(g.RepoURL, ".git"))
	cmd := fmt.Sprintf(`curl -X POST -H "Authorization: token %s" -H "Accept: application/vnd.github.v3+json" %s -d '{"title":"%s", "body":"%s", "head":"%s", "base":"%s"}'`,
		g.Token, prURL, title, description, sourceBranch, targetBranch)
	output, err := exec.Command("bash", "-c", cmd).Output()
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
