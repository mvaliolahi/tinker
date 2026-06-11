// Package run provides one-off Go code execution within the project context.
// It generates a temporary main.go that imports the project's packages,
// compiles and runs it, then cleans up.
package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const mainTemplate = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	{{- range .Imports}}
	{{.Alias}} "{{.Path}}"
	{{- end}}
)

func main() {
	ctx := context.Background()
	_ = ctx

	{{.Code}}
}
`

// Import represents a Go import declaration.
type Import struct {
	Alias string
	Path  string
}

// Config holds the configuration for a code execution.
type Config struct {
	ProjectModule string   // The project's Go module path (e.g., github.com/user/project)
	ProjectDir    string   // The project's root directory
	Code          string   // The Go code to execute
	Imports       []Import // Additional imports to include
}

// Runner executes one-off Go code within the project context.
type Runner struct {
	config Config
	tmpDir string
}

// NewRunner creates a new code runner.
func NewRunner(cfg Config) *Runner {
	return &Runner{config: cfg}
}

// Run generates, compiles, and executes the code.
func (r *Runner) Run() error {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "tinker-run-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	r.tmpDir = tmpDir
	defer r.cleanup()

	// Generate main.go
	if err := r.generateMain(); err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// Initialize go.mod in temp directory
	if err := r.initModule(); err != nil {
		return fmt.Errorf("initializing module: %w", err)
	}

	// Compile and run
	if err := r.compileAndRun(); err != nil {
		return fmt.Errorf("running code: %w", err)
	}

	return nil
}

// generateMain creates the temporary main.go file.
func (r *Runner) generateMain() error {
	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	data := struct {
		Imports []Import
		Code    string
	}{
		Imports: r.config.Imports,
		Code:    r.config.Code,
	}

	f, err := os.Create(filepath.Join(r.tmpDir, "main.go"))
	if err != nil {
		return fmt.Errorf("creating main.go: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// initModule sets up go.mod in the temp directory and adds a replace directive
// for the project module.
func (r *Runner) initModule() error {
	// Create go.mod
	cmd := exec.Command("go", "mod", "init", "tinker-run")
	cmd.Dir = r.tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init: %s: %w", string(output), err)
	}

	// Add replace directive to point to the actual project
	if r.config.ProjectModule != "" {
		cmd = exec.Command("go", "mod", "edit", "-replace",
			fmt.Sprintf("%s=%s", r.config.ProjectModule, r.config.ProjectDir))
		cmd.Dir = r.tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go mod edit -replace: %s: %w", string(output), err)
		}
	}

	// Tidy dependencies
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = r.tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %s: %w", string(output), err)
	}

	return nil
}

// compileAndRun builds and executes the temporary program.
func (r *Runner) compileAndRun() error {
	binaryPath := filepath.Join(r.tmpDir, "tinker-run")

	// Build
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = r.tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	// Run
	cmd = exec.Command(binaryPath)
	cmd.Dir = r.config.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// cleanup removes the temporary directory.
func (r *Runner) cleanup() {
	if r.tmpDir != "" {
		os.RemoveAll(r.tmpDir)
	}
}

// ParseCode extracts import statements and code from user input.
// Supports inline import comments: // import "package/path"
func ParseCode(input string) (code string, imports []Import) {
	var codeLines []string

	for _, line := range strings.Split(input, "\n") {
		trimmed := strings.TrimSpace(line)

		// Handle import comments: // import "package/path"
		if strings.HasPrefix(trimmed, "// import ") {
			pkgPath := strings.Trim(strings.TrimPrefix(trimmed, "// import "), `"`)
			imports = append(imports, Import{Path: pkgPath})
			continue
		}

		// Handle import statements
		if strings.HasPrefix(trimmed, "import ") {
			pkgPath := strings.Trim(strings.TrimPrefix(trimmed, "import "), `"`)
			imports = append(imports, Import{Path: pkgPath})
			continue
		}

		codeLines = append(codeLines, line)
	}

	code = strings.Join(codeLines, "\n")
	return
}
