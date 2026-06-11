package make

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func HasMakefile(dir string) bool {
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true
		}
	}
	return false
}

func Targets(dir string) []string {
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return parseTargets(data)
	}
	return nil
}

func parseTargets(data []byte) []string {
	var targets []string
	s := bufio.NewScanner(strings.NewReader(string(data)))
	for s.Scan() {
		line := s.Text()
		if len(line) == 0 || line[0] == '\t' || line[0] == ' ' {
			continue
		}
		if strings.HasPrefix(line, ".") || strings.Contains(line, "=") {
			continue
		}
		name, _, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		targets = append(targets, name)
	}
	return targets
}

func Run(dir, target string, args []string) error {
	if _, err := exec.LookPath("make"); err != nil {
		return fmt.Errorf("make not found — install GNU Make")
	}
	argv := append([]string{target}, args...)
	cmd := exec.Command("make", argv...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
