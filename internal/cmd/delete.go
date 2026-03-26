package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MTBorg/redwood/internal/git"
	"github.com/MTBorg/redwood/internal/tmux"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(baseDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <worktree>...",
		Short: "Delete one or more worktrees and their tmux sessions",
		Args:  cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
				if repo.Type != git.Worktree {
					continue
				}
				paths = append(paths, strings.Replace(repo.Path, home, "$HOME", 1))
			}
			return paths, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			for _, worktreePath := range args {
				worktreePath = strings.Replace(worktreePath, "$HOME", home, 1)

				if err := git.RemoveWorktree(worktreePath); err != nil {
					return fmt.Errorf("failed to remove worktree %q: %w", worktreePath, err)
				}

				sessionName := fmt.Sprintf("%s - %s", filepath.Base(filepath.Dir(worktreePath)), filepath.Base(worktreePath))
				if err := tmux.KillSession(sessionName); err != nil {
					// Non-fatal: session may not exist
					fmt.Fprintf(os.Stderr, "warning: could not kill tmux session %q: %v\n", sessionName, err)
				}
			}
			return nil
		},
	}
}
