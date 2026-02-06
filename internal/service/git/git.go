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

// GetCurrentBranch returns the current git branch name for the directory
// Returns an error if the directory is not a git repository or git command fails
func GetCurrentBranch(dir string) (string, error) {
	// Check if directory is a git repository
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	// Get current branch name
	cmd = exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// PullLatest pulls the latest changes from the remote repository
// Returns true if the pull brought new changes, false if already up-to-date
// Returns an error if the directory is not a git repository or git command fails
func PullLatest(dir string) (bool, error) {
	// Check if directory is a git repository
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("not a git repository: %w", err)
	}

	// Get the current HEAD commit before pull
	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	beforeOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get current HEAD: %w", err)
	}
	beforeCommit := strings.TrimSpace(string(beforeOutput))

	// Pull latest changes
	cmd = exec.Command("git", "-C", dir, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to pull: %w (output: %s)", err, string(output))
	}

	// Get the HEAD commit after pull
	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	afterOutput, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get current HEAD after pull: %w", err)
	}
	afterCommit := strings.TrimSpace(string(afterOutput))

	// If commits are different, there were updates
	hasUpdates := beforeCommit != afterCommit
	return hasUpdates, nil
}
