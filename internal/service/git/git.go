// Package git provides functionality for git operations
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// HasUncommittedChanges checks if the directory has uncommitted changes in git
// Returns true if there are uncommitted changes, false otherwise
// Returns an error if the directory is not a git repository or git command fails
func HasUncommittedChanges(dir string) (bool, error) {
	// Check if directory is a git repository
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("not a git repository: %w", err)
	}

	// Check git status - porcelain format for machine-readable output
	cmd = exec.Command("git", "-C", dir, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// If output is not empty, there are uncommitted changes
	status := strings.TrimSpace(string(output))
	return len(status) > 0, nil
}
