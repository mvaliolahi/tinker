package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mvaliolahi/tinker/internal/ui"
)

type Dep struct {
	Name     string
	Repo     string // Go repo path (e.g., "github.com/xo/usql")
	BuildDir string
	Purpose  string
	Pip      string // Python package name (e.g., "litecli") for pip-based installs
}

var all = []Dep{
	// Go-based tools
	{Name: "usql", Repo: "github.com/xo/usql", BuildDir: ".", Purpose: "database"},
	{Name: "grpcurl", Repo: "github.com/fullstorydev/grpcurl", BuildDir: "./cmd/grpcurl", Purpose: "grpc"},
	{Name: "evans", Repo: "github.com/ktr0731/evans", BuildDir: ".", Purpose: "grpc"},
	{Name: "curlie", Repo: "github.com/rs/curlie", BuildDir: ".", Purpose: "api"},
	// Python-based modern DB CLIs with syntax highlighting + autocomplete
	{Name: "litecli", Pip: "litecli", Purpose: "database"},
	{Name: "pgcli", Pip: "pgcli", Purpose: "database"},
	{Name: "mycli", Pip: "mycli", Purpose: "database"},
}

// installedCache caches IsInstalled results to avoid repeated filesystem lookups.
var (
	installedCache   = make(map[string]bool)
	installedCacheMu sync.RWMutex
)

func ForPurpose(purpose string) []Dep {
	var result []Dep
	for _, d := range all {
		if d.Purpose == purpose {
			result = append(result, d)
		}
	}
	return result
}

func All() []Dep { return all }

func IsInstalled(name string) bool {
	installedCacheMu.RLock()
	if cached, ok := installedCache[name]; ok {
		installedCacheMu.RUnlock()
		return cached
	}
	installedCacheMu.RUnlock()

	result := checkInstalled(name)

	installedCacheMu.Lock()
	installedCache[name] = result
	installedCacheMu.Unlock()

	return result
}

func checkInstalled(name string) bool {
	_, err := exec.LookPath(name)
	if err == nil {
		return true
	}
	// Check $HOME/go/bin/<name> (Go-installed binaries)
	p, _ := goBinPath(name)
	if p != "" {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	// Check $HOME/.local/bin/<name> (pip --user installed binaries)
	lp, _ := localBinPath(name)
	if lp != "" {
		if _, err := os.Stat(lp); err == nil {
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

	// Try pip install first for Python packages
	if dep.Pip != "" {
		if err := pipInstall(dep); err != nil {
			return fmt.Errorf("install %s: %s", dep.Name, err)
		}
		// Invalidate cache and recheck
		installedCacheMu.Lock()
		delete(installedCache, dep.Name)
		installedCacheMu.Unlock()
		if IsInstalled(dep.Name) {
			fmt.Printf("  %-10s %s\n", dep.Name, ui.Success("installed"))
			return nil
		}
		return fmt.Errorf("install %s: pip install succeeded but binary not found in PATH", dep.Name)
	}

	// Go-based: clone and build
	if dep.Repo != "" {
		if err := cloneAndBuild(dep); err != nil {
			return fmt.Errorf("install %s: %s", dep.Name, err)
		}
		// Invalidate cache and recheck
		installedCacheMu.Lock()
		delete(installedCache, dep.Name)
		installedCacheMu.Unlock()
		binPath, _ := goBinPath(dep.Name)
		if _, err := os.Stat(binPath); err != nil {
			return fmt.Errorf("install %s: binary not found after build", dep.Name)
		}
		fmt.Printf("  %-10s %s\n", dep.Name, ui.Success("installed"))
		return nil
	}

	return fmt.Errorf("install %s: no install method (neither pip nor go repo)", dep.Name)
}

func pipInstall(dep Dep) error {
	// Strategy 1: Try pipx (best for CLI tools — creates isolated venvs)
	if _, err := exec.LookPath("pipx"); err == nil {
		cmd := exec.Command("pipx", "install", dep.Pip)
		out, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		_ = out
	}

	// Find pip
	pipCmd := "pip3"
	if _, err := exec.LookPath("pip3"); err != nil {
		pipCmd = "pip"
		if _, err := exec.LookPath("pip"); err != nil {
			return fmt.Errorf("pip/pipx not found — install Python 3 or pipx to use %s", dep.Name)
		}
	}

	// Strategy 2: pip install --break-system-packages --user
	cmd := exec.Command(pipCmd, "install", "--break-system-packages", "--user", dep.Pip)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Strategy 3: Fallback without --break-system-packages (older pip)
	cmd2 := exec.Command(pipCmd, "install", "--user", dep.Pip)
	out2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return fmt.Errorf("pip install %s: %s", dep.Pip, shortErr(out2, err2))
	}
	_ = out
	return nil
}

func cloneAndBuild(dep Dep) error {
	// Strategy 1: Use "go install" (recommended, handles deps properly)
	installPath := dep.Repo
	if dep.BuildDir != "." && dep.BuildDir != "" {
		installPath = dep.Repo + "/" + strings.TrimPrefix(dep.BuildDir, "./")
	}
	cmd := exec.Command("go", "install", installPath+"@latest")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	// Strategy 2: Fallback to git clone + go build
	tmp, err2 := os.MkdirTemp("", "tinker-dep-*")
	if err2 != nil {
		return fmt.Errorf("go install: %s; git clone: %s", shortErr(out, err), err2)
	}
	defer os.RemoveAll(tmp)

	cloneURL := fmt.Sprintf("https://%s.git", dep.Repo)
	clone := exec.Command("git", "clone", "--depth", "1", "-q", cloneURL, tmp)
	if cout, err := clone.CombinedOutput(); err != nil {
		return fmt.Errorf("go install: %s; git clone: %s", shortErr(out, err), shortErr(cout, err))
	}

	binPath, err3 := goBinPath(dep.Name)
	if err3 != nil {
		return fmt.Errorf("go install: %s; go bin: %s", shortErr(out, err), err3)
	}

	build := exec.Command("go", "build", "-o", binPath, dep.BuildDir)
	build.Dir = tmp
	if bout, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("go install: %s; go build: %s", shortErr(out, err), shortErr(bout, err))
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

func localBinPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin", name), nil
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

// InstallAll installs all dependencies concurrently (Go and Python in parallel).
func InstallAll() (failed []Dep) {
	type result struct {
		dep Dep
		err error
	}

	goDeps := []Dep{}
	pipDeps := []Dep{}
	for _, dep := range All() {
		if dep.Pip != "" {
			pipDeps = append(pipDeps, dep)
		} else {
			goDeps = append(goDeps, dep)
		}
	}

	// Install Go and Python deps concurrently
	var wg sync.WaitGroup
	results := make(chan result, len(all))

	installBatch := func(deps []Dep) {
		for _, dep := range deps {
			wg.Add(1)
			go func(d Dep) {
				defer wg.Done()
				err := Install(d)
				results <- result{dep: d, err: err}
			}(dep)
		}
	}

	installBatch(goDeps)
	installBatch(pipDeps)

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			fmt.Printf("  %-10s %s\n", r.dep.Name, ui.Error(r.err.Error()))
			failed = append(failed, r.dep)
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
