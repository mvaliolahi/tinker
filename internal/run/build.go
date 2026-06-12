package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (r *Runner) initModule() error {
	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = r.tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s %v: %s: %w", name, args, string(out), err)
		}
		return nil
	}

	if err := run("go", "mod", "init", "tinker-run"); err != nil {
		return err
	}

	if r.cfg.ProjectModule != "" {
		if err := run("go", "mod", "edit", "-replace",
			fmt.Sprintf("%s=%s", r.cfg.ProjectModule, r.cfg.ProjectDir)); err != nil {
			return err
		}
	}

	return run("go", "mod", "tidy")
}

func (r *Runner) buildAndRun() error {
	bin := filepath.Join(r.tmpDir, "tinker-run")

	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = r.tmpDir
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	run := exec.Command(bin)
	run.Dir = r.cfg.ProjectDir
	run.Stdin, run.Stdout, run.Stderr = os.Stdin, os.Stdout, os.Stderr
	return run.Run()
}
