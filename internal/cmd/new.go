package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MTBorg/redwood/internal/config"
	"github.com/MTBorg/redwood/internal/git"
	"github.com/MTBorg/redwood/internal/tmux"
	"github.com/spf13/cobra"
)

func NewNewCmd(baseDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "new <repo> <worktree>",
		Short: "Create a new worktree in a git repository",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			base := *baseDir
			if base == "" {
				base = home
			}
			ignoredDirs := strings.Split(os.Getenv("REDWOOD_IGNORED_DIRS"), ",")
			repos, err := git.FindRepos(base, ignoredDirs)
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			paths := make([]string, 0, len(repos))
			for _, repo := range repos {
				paths = append(paths, strings.Replace(repo.Path, home, "$HOME", 1))
			}
			return paths, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, worktreeName := args[0], args[1]
			worktreePath, err := git.CreateWorktree(repoPath, worktreeName)
			if err != nil {
				return fmt.Errorf("failed to create worktree %q in %q: %w", worktreeName, repoPath, err)
			}
			sessionName := fmt.Sprintf("%s - %s", filepath.Base(repoPath), worktreeName)

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if err := tmux.NewSession(sessionName, worktreePath, cfg.Windows); err != nil {
				return fmt.Errorf("failed to create tmux session %q: %w", sessionName, err)
			}
			if err := tmux.AttachSession(sessionName); err != nil {
				return fmt.Errorf("failed to attach to tmux session %q: %w", sessionName, err)
			}
			return nil
		},
	}
}
