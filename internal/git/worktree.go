package git

import (
	"fmt"
	"os/exec"
)

// CreateWorktree creates a new worktree named worktreeName inside the
// repository at repoPath.
func CreateWorktree(repoPath, worktreeName string) error {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "add", worktreeName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add failed: %w\n%s", err, out)
	}
	return nil
}

// RemoveWorktree removes the worktree at worktreePath.
func RemoveWorktree(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "worktree", "remove", worktreePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove failed: %w\n%s", err, out)
	}
	return nil
}
