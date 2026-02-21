package cmd

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/MTBorg/redwood/internal/tmux"
	"github.com/spf13/cobra"
)

func NewOpenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <path>",
		Short: "Open a git repository in a new tmux session",
		Args:  cobra.ExactArgs(1),
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
