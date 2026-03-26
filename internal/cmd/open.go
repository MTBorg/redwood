package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/MTBorg/redwood/internal/git"
	"github.com/MTBorg/redwood/internal/tmux"
	"github.com/spf13/cobra"
)

func NewOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <path>",
		Short: "Open a git repository in a new tmux session",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			ignoredDirs := strings.Split(os.Getenv("REDWOOD_IGNORED_DIRS"), ",")
			repos, err := git.FindRepos(home, ignoredDirs)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			paths := make([]string, len(repos))
			for i, repo := range repos {
				paths[i] = strings.Replace(repo.Path, home, "$HOME", 1)
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

			if err := tmux.NewSession(sessionName, path); err != nil {
				return fmt.Errorf("failed to create tmux session %q: %w", sessionName, err)
			}

			if err := tmux.AttachSession(sessionName); err != nil {
				return fmt.Errorf("failed to attach to tmux session %q: %w", sessionName, err)
			}

			return nil
		},
	}
}
