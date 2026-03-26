package tmux

import (
	"os"
	"os/exec"
	"path/filepath"
)

func configPath() string {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		cfg = filepath.Join(home, ".config")
	}
	return filepath.Join(cfg, "tmux", ".tmux.conf")
}

// InSession reports whether the current process is running inside a tmux session.
func InSession() bool {
	return os.Getenv("TMUX") != ""
}

// NewSession creates a new detached tmux session with the given name, using dir
// as the start directory. The user's tmux config is loaded if it exists.
func NewSession(name, dir string) error {
	args := []string{"new-session", "-d", "-s", name, "-c", dir}
	if cfg := configPath(); cfg != "" {
		if _, err := os.Stat(cfg); err == nil {
			args = append([]string{"-f", cfg}, args...)
		}
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// KillSession kills the tmux session with the given name.
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AttachSession attaches to an existing tmux session. If already inside tmux,
// it switches the client to the target session instead of nesting.
func AttachSession(name string) error {
	var cmd *exec.Cmd
	if InSession() {
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", name)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
