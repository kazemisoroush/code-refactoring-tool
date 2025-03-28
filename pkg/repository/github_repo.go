// Package repository provides a generic interface for interacting with code repositories.
package repository

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitHubRepo represents a GitHub repository
type GitHubRepo struct {
	RepoURL string
	Token   string
	repo    *git.Repository
	path    string
}

// NewGitHubRepo creates a new GitHub repository instance
func NewGitHubRepo(repoURL, token string) Repository {
	repoName := path.Base(strings.TrimSuffix(repoURL, ".git"))
	return &GitHubRepo{
		RepoURL: repoURL,
		Token:   token,
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
		return err
	}
	g.repo = repo
	return nil
}

// CheckoutBranch creates a new branch and checks it out
func (g *GitHubRepo) CheckoutBranch(branchName string) error {
	wt, err := g.repo.Worktree()
	if err != nil {
		return err
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
		return err
	}
	if _, err := wt.Add("."); err != nil {
		return err
	}
	_, err = wt.Commit(message, &git.CommitOptions{})
	return err
}

// Push pushes commits to the remote repository
func (g *GitHubRepo) Push() error {
	return g.repo.Push(&git.PushOptions{})
}

// CreatePR creates a new pull request (still using curl for simplicity)
func (g *GitHubRepo) CreatePR(title, description, sourceBranch, targetBranch string) (string, error) {
	prURL := fmt.Sprintf("https://api.github.com/repos/%s/pulls", "OWNER/REPO")
	cmd := fmt.Sprintf(`curl -X POST -H "Authorization: token %s" -H "Accept: application/vnd.github.v3+json" %s -d '{"title":"%s", "body":"%s", "head":"%s", "base":"%s"}'`,
		g.Token, prURL, title, description, sourceBranch, targetBranch)
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Cleanup deletes the repository from the filesystem.
func (g *GitHubRepo) Cleanup() error {
	return os.RemoveAll(g.path)
}
