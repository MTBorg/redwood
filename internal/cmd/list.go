package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/MTBorg/redwood/internal/git"
	"github.com/spf13/cobra"
)

func NewListCmd(includedDirs *[]string) *cobra.Command {
	var onlyBareRepos bool
	var onlyWorktrees bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all git repositories in your home directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			bases := *includedDirs
			if len(bases) == 0 {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("could not determine home directory: %w", err)
				}
				bases = []string{home}
			}

			ignoredDirs := strings.Split(os.Getenv("REDWOOD_IGNORED_DIRS"), ",")
			slog.Debug("searching for git repos", "bases", bases, "ignored", ignoredDirs)

			var repos []git.Repo
			for _, base := range bases {
				r, err := git.FindRepos(expandTilde(base), ignoredDirs)
				if err != nil {
					return fmt.Errorf("error searching for repositories: %w", err)
				}
				repos = append(repos, r...)
			}

			home, _ := os.UserHomeDir()
			for _, repo := range repos {
				if onlyBareRepos && repo.Type != git.Bare {
					continue
				}
				if onlyWorktrees && repo.Type != git.Worktree {
					continue
				}
				path := strings.Replace(repo.Path, home, "$HOME", 1)
				fmt.Println(path)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&onlyBareRepos, "only-bare-repos", false, "only list bare repositories")
	cmd.Flags().BoolVar(&onlyWorktrees, "only-worktrees", false, "only list worktrees")

	return cmd
}
