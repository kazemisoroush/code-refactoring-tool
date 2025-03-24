package fixer

// Fixer is the interface that wraps the Fix method.
//
//go:generate mockgen -destination=./mocks/mock_fixer.go -mock_names=Fixer=MockFixer -package=mocks . Fixer
type Fixer interface {
	// FixIssues fixes the code in the provided source path
	FixIssues(sourcePath string) error
}
