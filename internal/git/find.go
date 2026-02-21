package git

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type RepoType int

const (
	Regular  RepoType = iota
	Bare
	Worktree
)

type Repo struct {
	Path string
	Type RepoType
}

func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// detectRepoType checks whether path is a git repository and returns its type.
// Returns (type, true) if it's a git repo, or (Regular, false) if it's not.
func detectRepoType(path string) (RepoType, bool) {
	gitPath := filepath.Join(path, ".git")
	fi, err := os.Stat(gitPath)
	if err == nil {
		if fi.IsDir() {
			slog.Debug("found regular git repo", "path", path)
			return Regular, true
		}
		// .git is a file => linked worktree
		slog.Debug("found git worktree", "path", path)
		return Worktree, true
	}

	// Check for bare repo: HEAD + objects/ + refs/ at the root
	_, headErr := os.Stat(filepath.Join(path, "HEAD"))
	_, objErr := os.Stat(filepath.Join(path, "objects"))
	_, refsErr := os.Stat(filepath.Join(path, "refs"))
	if headErr == nil && objErr == nil && refsErr == nil {
		slog.Debug("found bare git repo", "path", path)
		return Bare, true
	}

	return Regular, false
}

// FindRepos recursively searches root for git repositories, skipping hidden
// directories, symlinks, and any directory named in ignoredDirs.
// The "worktrees" directory is always skipped to avoid git's internal worktree
// metadata (bare repos store linked worktree metadata in <repo>/worktrees/).
func FindRepos(root string, ignoredDirs []string) ([]Repo, error) {
	ignored := map[string]bool{"worktrees": true}
	for _, d := range ignoredDirs {
		if d != "" {
			ignored[d] = true
		}
	}

	var repos []Repo

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				slog.Debug("skipping directory (permission denied)", "path", path)
				return filepath.SkipDir
			}
			return err
		}

		if !d.IsDir() {
			return nil
		}

		name := d.Name()

		if path != root {
			if isSymlink(path) {
				slog.Debug("skipping symlink", "path", path)
				return filepath.SkipDir
			}
			if isHidden(name) {
				slog.Debug("skipping hidden directory", "path", path)
				return filepath.SkipDir
			}
			if ignored[name] {
				slog.Debug("skipping ignored directory", "path", path)
				return filepath.SkipDir
			}
		}

		repoType, isRepo := detectRepoType(path)
		if isRepo {
			repos = append(repos, Repo{Path: path, Type: repoType})
		}

		return nil
	})

	return repos, err
}
