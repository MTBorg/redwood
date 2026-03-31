package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateWorktree creates a new worktree named worktreeName as a sibling of the
// repository at repoPath (i.e. ../repoName-worktreeName) and returns its path.
func CreateWorktree(repoPath, worktreeName string) (string, error) {
	worktreePath := filepath.Join(filepath.Dir(repoPath), filepath.Base(repoPath)+"-"+worktreeName)
	cmd := exec.Command("git", "-C", repoPath, "worktree", "add", worktreePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree add failed: %w\n%s", err, out)
	}
	return worktreePath, nil
}

// ListWorktrees returns the paths of all worktrees associated with the
// repository at repoPath, excluding the main worktree itself.
func ListWorktrees(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var paths []string
	first := true
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if first {
				first = false
				continue // skip main worktree
			}
			paths = append(paths, path)
		}
	}
	return paths, scanner.Err()
}

// RemoveWorktree removes the worktree at worktreePath.
func RemoveWorktree(worktreePath string) error {
	cmd := exec.Command("git", "-C", worktreePath, "worktree", "remove", worktreePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove failed: %w\n%s", err, out)
	}
	return nil
}
