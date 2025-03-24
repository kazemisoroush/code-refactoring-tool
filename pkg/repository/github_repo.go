package repository

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

type GitHubRepo struct {
	RepoURL string
	Token   string
}

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

func (g *GitHubRepo) Clone() error {
	cmd := exec.Command("git", "clone", g.RepoURL)
	return cmd.Run()
}

func (g *GitHubRepo) CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	return cmd.Run()
}

func (g *GitHubRepo) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-am", message)
	return cmd.Run()
}

func (g *GitHubRepo) Push() error {
	cmd := exec.Command("git", "push", "origin", "HEAD")
	return cmd.Run()
}

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
