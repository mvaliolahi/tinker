package runner

import (
	"os"
	"os/exec"

	"golang.org/x/term"
)

func Interactive(name string, args ...string) error {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err == nil {
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), state) }()
	}

	cmd := exec.Command(name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
