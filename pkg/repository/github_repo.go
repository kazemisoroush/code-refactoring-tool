// Package repository provides a generic interface for interacting with code repositories.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
func (g *GitHubRepo) Clone(ctx context.Context) error {
	repo, err := git.PlainCloneContext(ctx, g.path, false, &git.CloneOptions{
		URL:      g.RepoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		slog.Error("something went wrong", "error", err)
		return nil
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
func (g *GitHubRepo) Push(ctx context.Context) error {
	return g.repo.PushContext(ctx, &git.PushOptions{
		RemoteURL:  g.RepoURL,
		RemoteName: "origin",
		Force:      true,
		Auth: &gitHttp.BasicAuth{
			Username: g.Author,
			Password: g.Token,
		},
	})
}

// CreatePR creates a new pull request
func (g *GitHubRepo) CreatePR(title, description, sourceBranch, targetBranch string) (string, error) {
	owner, repo, err := g.getOwnerRepo()
	if err != nil {
		return "", err
	}

	createURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	cmd := fmt.Sprintf(`curl -X POST -H "Authorization: token %s" -H "Accept: application/vnd.github.v3+json" %s -d '{"title":"%s", "body":"%s", "head":"%s", "base":"%s"}'`,
		g.Token, createURL, title, description, sourceBranch, targetBranch)
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("failed to create pull request: %w", err)
	}
	return string(output), nil
}

// UpsertPR creates a PR if it doesn't exist, otherwise updates the existing one.
func (g *GitHubRepo) UpsertPR(title, description, sourceBranch, targetBranch string) (string, error) {
	exists, prNumber, err := g.PRExists(sourceBranch, targetBranch)
	if err != nil {
		return "", err
	}
	if exists {
		if err := g.UpdatePR(prNumber, title, description); err != nil {
			return "", err
		}
		return fmt.Sprintf("PR #%d updated", prNumber), nil
	}
	return g.CreatePR(title, description, sourceBranch, targetBranch)
}

// PRExists checks if a PR exists and returns (true, PR number) or (false, 0)
func (g *GitHubRepo) PRExists(sourceBranch, targetBranch string) (bool, int, error) {
	owner, repo, err := g.getOwnerRepo()
	if err != nil {
		return false, 0, err
	}

	listURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?head=%s:%s&base=%s", owner, repo, owner, sourceBranch, targetBranch)
	cmd := fmt.Sprintf(`curl -s -H "Authorization: token %s" %s`, g.Token, listURL)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false, 0, fmt.Errorf("failed to check existing PRs: %w", err)
	}

	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(out, &prs); err != nil {
		return false, 0, fmt.Errorf("failed to parse PR list: %w", err)
	}

	if len(prs) > 0 {
		return true, prs[0].Number, nil
	}
	return false, 0, nil
}

// UpdatePR updates an existing pull request's title and body
func (g *GitHubRepo) UpdatePR(prNumber int, title, description string) error {
	owner, repo, err := g.getOwnerRepo()
	if err != nil {
		return err
	}

	updateURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNumber)
	cmd := fmt.Sprintf(`curl -X PATCH -H "Authorization: token %s" -H "Accept: application/vnd.github.v3+json" %s -d '{"title":"%s", "body":"%s"}'`,
		g.Token, updateURL, title, description)
	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}
	return nil
}

// Cleanup deletes the repository from the filesystem.
func (g *GitHubRepo) Cleanup() error {
	err := os.RemoveAll(g.path)
	if err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}
	return nil
}

// getOwnerRepo extracts owner and repo from the GitHub URL
func (g *GitHubRepo) getOwnerRepo() (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(g.RepoURL, "https://github.com/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub repo URL: %s", g.RepoURL)
	}
	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")
	return owner, repo, nil
}
