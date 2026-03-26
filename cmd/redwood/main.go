package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/MTBorg/redwood/internal/cmd"
	"github.com/spf13/cobra"
)

func setupLogging(logFile string) error {
	var handler slog.Handler
	if logFile == "stderr" {
		handler = slog.NewTextHandler(os.Stderr, nil)
	} else {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			handler = slog.NewTextHandler(io.Discard, nil)
		} else {
			handler = slog.NewTextHandler(f, nil)
		}
	}
	slog.SetDefault(slog.New(handler))
	return nil
}

func main() {
	var logFile string
	var baseDir string

	root := &cobra.Command{
		Use:   "redwood2",
		Short: "Integrate git repositories with tmux sessions",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			return setupLogging(logFile)
		},
	}

	root.PersistentFlags().StringVar(&logFile, "log-file", "/var/log/redwood.log", `log file path (use "stderr" to write to stderr)`)
	root.PersistentFlags().StringVar(&baseDir, "base", "", "base directory to search (defaults to home directory)")

	root.AddCommand(cmd.CompletionCmd)
	root.AddCommand(cmd.NewListCmd(&baseDir))
	root.AddCommand(cmd.NewOpenCmd(&baseDir))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
