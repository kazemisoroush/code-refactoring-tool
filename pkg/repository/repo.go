package repository

import "context"

// Repository is an interface for interacting with a git repository
//
//go:generate mockgen -destination=./mocks/mock_repository.go -mock_names=Repository=MockRepository -package=mocks . Repository
type Repository interface {
	// Clone clones a git repository
	Clone(ctx context.Context) error

	// GetPath returns the path to the repository
	GetPath() string

	// CheckoutBranch checks out a branch in the repository
	CheckoutBranch(branchName string) error

	// Commit commits changes in the repository
	Commit(message string) error

	// Push pushes the current branch to the remote repository
	Push() error

	// UpsertPR creates a PR if not exists otherwise updates the existing PR
	UpsertPR(title, description, sourceBranch, targetBranch string) (string, error)

	// CreatePR creates a pull request in the repository
	CreatePR(title, description, sourceBranch, targetBranch string) (string, error)

	// UpdatePR updates a pull request in the repository
	UpdatePR(prNumber int, title, description string) error

	// PRExists checks if a PR exists in the repository
	PRExists(sourceBranch, targetBranch string) (bool, int, error)

	// Cleanup deletes the repository
	Cleanup() error
}
