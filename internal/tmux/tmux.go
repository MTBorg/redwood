package tmux

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MTBorg/redwood/internal/config"
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

// SessionExists reports whether a tmux session with the given name already exists.
func SessionExists(name string) bool {
	return exec.Command("tmux", "has-session", "-t", name).Run() == nil
}

// NewSession creates a new detached tmux session with the given name, using dir
// as the start directory. The user's tmux config is loaded if it exists.
// If windows is non-empty, named windows are created with optional startup commands.
func NewSession(name, dir string, windows []config.Window) error {
	firstWindowName := ""
	if len(windows) > 0 {
		firstWindowName = windows[0].Name
	}

	args := []string{"new-session", "-d", "-s", name, "-c", dir}
	if firstWindowName != "" {
		args = append(args, "-n", firstWindowName)
	}
	if cfg := configPath(); cfg != "" {
		if _, err := os.Stat(cfg); err == nil {
			args = append([]string{"-f", cfg}, args...)
		}
	}
	cmd := exec.Command("tmux", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Send command to first window if set
	if len(windows) > 0 && windows[0].Command != "" {
		if err := sendKeys(name+":"+windows[0].Name, windows[0].Command); err != nil {
			return err
		}
	}

	// Create remaining windows
	for _, w := range windows[1:] {
		newWinArgs := []string{"new-window", "-t", name, "-n", w.Name, "-c", dir}
		if err := exec.Command("tmux", newWinArgs...).Run(); err != nil {
			return err
		}
		if w.Command != "" {
			if err := sendKeys(name+":"+w.Name, w.Command); err != nil {
				return err
			}
		}
	}

	for _, w := range windows {
		if w.Focus {
			if err := exec.Command("tmux", "select-window", "-t", name+":"+w.Name).Run(); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func sendKeys(target, command string) error {
	return exec.Command("tmux", "send-keys", "-t", target, command, "Enter").Run()
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
