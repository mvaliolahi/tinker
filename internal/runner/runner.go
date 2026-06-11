package runner

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty/v2"
	"golang.org/x/term"
)

func Interactive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		return cmd.Run()
	}
	defer ptmx.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err == nil {
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	pty.InheritSize(os.Stdin, ptmx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			pty.InheritSize(os.Stdin, ptmx)
		}
		signal.Stop(ch)
	}()

	go relayOut(ptmx)
	go relayIn(ptmx)

	return cmd.Wait()
}

func relayOut(ptmx *os.File) {
	buf := make([]byte, 8192)
	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}

func relayIn(ptmx *os.File) {
	buf := make([]byte, 8192)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				if buf[i] == '\r' {
					buf[i] = '\n'
				}
			}
			ptmx.Write(buf[:n])
		}
		if err != nil {
			return
		}
	}
}
