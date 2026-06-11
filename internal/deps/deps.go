package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mvaliolahi/tinker/internal/ui"
)

type Dep struct {
	Name     string
	Repo     string
	BuildDir string
	Purpose  string
}

var all = []Dep{
	{Name: "usql", Repo: "github.com/xo/usql", BuildDir: ".", Purpose: "database"},
	{Name: "grpcurl", Repo: "github.com/fullstorydev/grpcurl", BuildDir: "./cmd/grpcurl", Purpose: "grpc"},
	{Name: "evans", Repo: "github.com/ktr0731/evans", BuildDir: ".", Purpose: "grpc"},
	{Name: "curlie", Repo: "github.com/rs/curlie", BuildDir: ".", Purpose: "api"},
}

func ForPurpose(purpose string) []Dep {
	var result []Dep
	for _, d := range all {
		if d.Purpose == purpose {
			result = append(result, d)
		}
	}
	return result
}

func All() []Dep {
	return all
}

func IsInstalled(name string) bool {
	_, err := exec.LookPath(name)
	if err == nil {
		return true
	}
	p, _ := goBinPath(name)
	if p != "" {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func Install(dep Dep) error {
	if IsInstalled(dep.Name) {
		fmt.Printf("  %-10s %s\n", dep.Name, ui.Success("installed"))
		return nil
	}

	fmt.Printf("  %-10s %s\n", dep.Name, ui.Accent("installing..."))

	if err := cloneAndBuild(dep); err != nil {
		return fmt.Errorf("install %s: %s", dep.Name, err)
	}

	binPath, _ := goBinPath(dep.Name)
	if _, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("install %s: binary not found after build", dep.Name)
	}

	fmt.Printf("  %-10s %s\n", dep.Name, ui.Success("installed"))
	return nil
}

func cloneAndBuild(dep Dep) error {
	tmp, err := os.MkdirTemp("", "tinker-dep-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	cloneURL := fmt.Sprintf("https://%s.git", dep.Repo)
	clone := exec.Command("git", "clone", "--depth", "1", "-q", cloneURL, tmp)
	if out, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %s", shortErr(out, err))
	}

	binPath, err := goBinPath(dep.Name)
	if err != nil {
		return err
	}

	build := exec.Command("go", "build", "-o", binPath, dep.BuildDir)
	build.Dir = tmp
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %s", shortErr(out, err))
	}

	return nil
}

func goBinPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "go", "bin", name), nil
}

func shortErr(out []byte, err error) string {
	s := strings.TrimSpace(string(out))
	if len(s) > 120 {
		s = s[:120] + "..."
	}
	if s != "" {
		return s + ": " + err.Error()
	}
	return err.Error()
}

func InstallForPurpose(purpose string) (failed []Dep) {
	for _, dep := range ForPurpose(purpose) {
		if err := Install(dep); err != nil {
			fmt.Printf("  %-10s %s\n", dep.Name, ui.Error(err.Error()))
			failed = append(failed, dep)
		}
	}
	return failed
}

func InstallAll() (failed []Dep) {
	for _, dep := range All() {
		if err := Install(dep); err != nil {
			fmt.Printf("  %-10s %s\n", dep.Name, ui.Error(err.Error()))
			failed = append(failed, dep)
		}
	}
	return failed
}

func Check() (missing []Dep) {
	for _, dep := range All() {
		if !IsInstalled(dep.Name) {
			missing = append(missing, dep)
		}
	}
	return missing
}

func FormatList(deps []Dep) string {
	var names []string
	for _, d := range deps {
		names = append(names, d.Name)
	}
	return strings.Join(names, ", ")
}
