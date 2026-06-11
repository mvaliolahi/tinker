package runner

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty/v2"
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

	pty.InheritSize(os.Stdin, ptmx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			pty.InheritSize(os.Stdin, ptmx)
		}
		signal.Stop(ch)
	}()

	go func() { _, _ = copyToStdout(ptmx) }()
	go func() { _, _ = copyFromStdin(ptmx) }()

	return cmd.Wait()
}

func copyToStdout(ptmx *os.File) (int, error) {
	buf := make([]byte, 8192)
	var total int
	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
			total += n
		}
		if err != nil {
			return total, err
		}
	}
}

func copyFromStdin(ptmx *os.File) (int, error) {
	buf := make([]byte, 8192)
	var total int
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			ptmx.Write(buf[:n])
			total += n
		}
		if err != nil {
			return total, err
		}
	}
}
