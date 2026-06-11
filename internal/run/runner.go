package run

import (
	"os"
	"path/filepath"
	"text/template"
)

type Runner struct {
	cfg    Config
	tmpDir string
}

func NewRunner(cfg Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Run() error {
	tmp, err := os.MkdirTemp("", "tinker-run-*")
	if err != nil {
		return err
	}
	r.tmpDir = tmp
	defer os.RemoveAll(tmp)

	if err := r.writeMain(); err != nil {
		return err
	}
	if err := r.initModule(); err != nil {
		return err
	}
	return r.buildAndRun()
}

func (r *Runner) writeMain() error {
	tmpl, err := template.New("main").Parse(mainTmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(r.tmpDir, "main.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, struct {
		Imports []Import
		Code    string
	}{r.cfg.Imports, r.cfg.Code})
}
