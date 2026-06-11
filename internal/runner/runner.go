package runner

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/term"
)

func Interactive(name string, args ...string) error {
	// Save terminal state so we can restore after child exits
	// (in case the child crashes without restoring it)
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err == nil {
		defer term.Restore(int(os.Stdin.Fd()), state)
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}
	return cmd.Run()
}
