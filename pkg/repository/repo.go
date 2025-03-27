package repository

// Repository is an interface for interacting with a git repository
//
//go:generate mockgen -destination=./mocks/mock_repository.go -mock_names=Repository=MockRepository -package=mocks . Repository
type Repository interface {
	// Clone clones a git repository
	Clone() error

	// GetPath returns the path to the repository
	GetPath() string

	// CheckoutBranch checks out a branch in the repository
	CheckoutBranch(branchName string) error

	// Commit commits changes in the repository
	Commit(message string) error

	// Push pushes the current branch to the remote repository
	Push() error

	// CreatePR creates a pull request in the repository
	CreatePR(title, description, sourceBranch, targetBranch string) (string, error)

	// Cleanup deletes the repository
	Cleanup() error
}
