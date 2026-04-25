package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/MTBorg/redwood/internal/config"
	"github.com/MTBorg/redwood/internal/git"
	"github.com/MTBorg/redwood/internal/tmux"
	"github.com/spf13/cobra"
)

func NewOpenCmd(includedDirs *[]string) *cobra.Command {
	return &cobra.Command{
		Use:   "open <path>",
		Short: "Open a git repository in a new tmux session",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			bases := *includedDirs
			if len(bases) == 0 {
				bases = []string{home}
			}
			ignoredDirs := strings.Split(os.Getenv("REDWOOD_IGNORED_DIRS"), ",")
			var repos []git.Repo
			for _, base := range bases {
				r, err := git.FindRepos(expandTilde(base), ignoredDirs)
				if err != nil {
					return nil, cobra.ShellCompDirectiveError
				}
				repos = append(repos, r...)
			}

			seen := make(map[string]bool)
			var paths []string
			addPath := func(p string) {
				if !seen[p] {
					seen[p] = true
					paths = append(paths, strings.Replace(p, home, "$HOME", 1))
				}
			}

			for _, repo := range repos {
				addPath(repo.Path)
				worktrees, err := git.ListWorktrees(repo.Path)
				if err != nil {
					slog.Debug("failed to list worktrees", "repo", repo.Path, "err", err)
					continue
				}
				for _, wt := range worktrees {
					addPath(wt)
				}
			}
			return paths, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Canonicalize: resolve relative paths and symlinks
			path, err := filepath.EvalSymlinks(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", args[0], err)
			}
			path, err = filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", args[0], err)
			}

			sessionName := filepath.Base(path)
			slog.Debug("opening repo in tmux", "path", path, "session", sessionName)

			if !tmux.SessionExists(sessionName) {
				cfg, err := config.Load()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				if err := tmux.NewSession(sessionName, path, cfg.Windows); err != nil {
					return fmt.Errorf("failed to create tmux session %q: %w", sessionName, err)
				}
			} else {
				slog.Debug("tmux session already exists, attaching", "session", sessionName)
			}

			if err := tmux.AttachSession(sessionName); err != nil {
				return fmt.Errorf("failed to attach to tmux session %q: %w", sessionName, err)
			}

			return nil
		},
	}
}
