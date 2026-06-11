package deps

import (
	"fmt"
	"os"
	"os/exec"
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

func Install(purpose string) {
	for _, dep := range all {
		if dep.Purpose != purpose {
			continue
		}

		if _, err := exec.LookPath(dep.Name); err == nil {
			fmt.Printf("  %-10s ✓ already installed\n", dep.Name)
			continue
		}

		fmt.Printf("  %-10s installing...\n", dep.Name)
		cmd := exec.Command("go", "install", dep.Module)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("  %-10s ✗ failed (run manually: go install %s)\n", dep.Name, dep.Module)
		} else {
			fmt.Printf("  %-10s ✓ installed\n", dep.Name)
		}
	}
}
