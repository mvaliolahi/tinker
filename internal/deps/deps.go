package deps

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Dep struct {
	Name    string
	Module  string
	Purpose string
}

var all = []Dep{
	{Name: "usql", Module: "github.com/xo/usql@latest", Purpose: "database"},
	{Name: "grpcurl", Module: "github.com/fullstorydev/grpcurl/cmd/grpcurl@latest", Purpose: "grpc"},
	{Name: "evans", Module: "github.com/ktr0731/evans@latest", Purpose: "grpc"},
	{Name: "curlie", Module: "github.com/rs/curlie@latest", Purpose: "api"},
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
	return err == nil
}

func Install(dep Dep) error {
	if IsInstalled(dep.Name) {
		fmt.Printf("  %-10s ✓ already installed\n", dep.Name)
		return nil
	}

	fmt.Printf("  %-10s installing...\n", dep.Name)

	if err := goInstall(dep.Module, false); err != nil {
		fmt.Printf("  %-10s proxy failed, retrying with GOPROXY=direct...\n", dep.Name)
		if err2 := goInstall(dep.Module, true); err2 != nil {
			return fmt.Errorf("install %s: run manually: go install %s", dep.Name, dep.Module)
		}
	}

	fmt.Printf("  %-10s ✓ installed\n", dep.Name)
	return nil
}

func goInstall(module string, direct bool) error {
	cmd := exec.Command("go", "install", module)
	env := os.Environ()
	if direct {
		env = append(env, "GOPROXY=direct")
	}
	cmd.Env = env
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func InstallForPurpose(purpose string) (failed []Dep) {
	for _, dep := range ForPurpose(purpose) {
		if err := Install(dep); err != nil {
			fmt.Printf("  %-10s ✗ %s\n", dep.Name, err)
			failed = append(failed, dep)
		}
	}
	return failed
}

func InstallAll() (failed []Dep) {
	for _, dep := range All() {
		if err := Install(dep); err != nil {
			fmt.Printf("  %-10s ✗ %s\n", dep.Name, err)
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
