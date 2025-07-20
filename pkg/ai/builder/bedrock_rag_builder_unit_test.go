package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBedrockRAGBuilder_getRDSTableName tests the private getRDSTableName method
// to ensure it properly sanitizes repository paths into valid SQL identifiers.
func TestBedrockRAGBuilder_getRDSTableName(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "simple repo name",
			repoPath: "my-project",
			expected: "my_project",
		},
		{
			name:     "repo name with hyphens",
			repoPath: "code-refactoring-test",
			expected: "code_refactoring_test",
		},
		{
			name:     "full path with special characters",
			repoPath: "/path/to/my-repo@v1.0",
			expected: "my_repo_v1_0",
		},
		{
			name:     "starts with number",
			repoPath: "123-project",
			expected: "_123_project",
		},
		{
			name:     "already valid name",
			repoPath: "valid_repo_name",
			expected: "valid_repo_name",
		},
		{
			name:     "mixed case with special chars",
			repoPath: "MyRepo-V2.1",
			expected: "myrepo_v2_1",
		},
		{
			name:     "path with directories",
			repoPath: "/home/user/projects/awesome-project",
			expected: "awesome_project",
		},
		{
			name:     "relative path",
			repoPath: "projects/my-app",
			expected: "my_app",
		},
		{
			name:     "dots and spaces",
			repoPath: "my.project name",
			expected: "my_project_name",
		},
		{
			name:     "starts with underscore",
			repoPath: "_private-repo",
			expected: "_private_repo",
		},
		{
			name:     "multiple consecutive special chars",
			repoPath: "test---repo___name",
			expected: "test___repo___name",
		},
		{
			name:     "empty base name edge case",
			repoPath: "/",
			expected: "_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a BedrockRAGBuilder instance with the test repo path
			builder := &BedrockRAGBuilder{
				repoPath: tt.repoPath,
				// Other fields are not needed for this test
			}

			// Call the private method
			result := builder.getRDSTableName()

			// Assert the result
			assert.Equal(t, tt.expected, result,
				"getRDSTableName() returned unexpected result for repoPath: %s", tt.repoPath)

			// Additional assertions to ensure the result is a valid SQL identifier
			assert.NotEmpty(t, result, "Result should not be empty")

			// Check that it starts with letter or underscore
			if len(result) > 0 {
				firstChar := result[0]
				assert.True(t,
					(firstChar >= 'a' && firstChar <= 'z') ||
						(firstChar >= 'A' && firstChar <= 'Z') ||
						firstChar == '_',
					"Result should start with letter or underscore, got: %c", firstChar)
			}

			// Check that all characters are valid SQL identifier characters
			for i, char := range result {
				valid := (char >= 'a' && char <= 'z') ||
					(char >= 'A' && char <= 'Z') ||
					(char >= '0' && char <= '9') ||
					char == '_'
				assert.True(t, valid,
					"Character at position %d (%c) is not a valid SQL identifier character", i, char)
			}
		})
	}
}

func TestBedrockRAGBuilder_getRDSTableName_SQLCompliance(t *testing.T) {
	// Additional test to verify SQL compliance with some edge cases
	edgeCases := []struct {
		name     string
		repoPath string
	}{
		{"unicode characters", "项目-测试"},
		{"emojis", "my-project-🚀"},
		{"very long name", "this-is-a-very-long-repository-name-that-might-cause-issues-in-some-databases"},
		{"only numbers", "12345"},
		{"only special chars", "---...___"},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := &BedrockRAGBuilder{repoPath: tc.repoPath}
			result := builder.getRDSTableName()

			// Ensure result is not empty
			assert.NotEmpty(t, result, "Result should not be empty for edge case: %s", tc.name)

			// Ensure it's a valid PostgreSQL identifier
			// - Must start with letter or underscore
			// - Can contain letters, digits, underscores
			// - Case insensitive (we convert to lowercase)
			assert.Regexp(t, `^[a-z_][a-z0-9_]*$`, result,
				"Result should match PostgreSQL identifier pattern for edge case: %s", tc.name)
		})
	}
}
