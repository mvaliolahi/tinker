package grpc

import (
	"fmt"
	"os/exec"

	"github.com/mvaliolahi/tinker/internal/runner"
)

func (s *Session) baseArgs() []string {
	args := []string{"-plaintext"}
	if !s.Reflection && s.ProtoDir != "" {
		args = append(args, "-import-path", s.ProtoDir, "-proto", s.ProtoDir+"/...")
	}
	return args
}

// ListServices lists all gRPC services. Uses native Go gRPC client first
// (no external binary needed when server reflection is enabled), then
// falls back to the grpcurl CLI.
func (s *Session) ListServices() (string, error) {
	if s.Reflection {
		native := NewNativeClient(s.Addr, s.ProtoDir, s.Reflection)
		if out, err := native.ListServices(); err == nil {
			return out, nil
		}
	}
	return s.listServicesCLI()
}

func (s *Session) listServicesCLI() (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found and native gRPC reflection failed — run 'tinker deps' to install")
	}

	args := append(s.baseArgs(), s.Addr, "list")
	out, err := exec.Command("grpcurl", args...).CombinedOutput() //nolint:gosec // args built from our own config
	return string(out), err
}

// Describe describes a gRPC service. Uses native client first, then CLI fallback.
func (s *Session) Describe(service string) (string, error) {
	if s.Reflection {
		native := NewNativeClient(s.Addr, s.ProtoDir, s.Reflection)
		if out, err := native.Describe(service); err == nil {
			return out, nil
		}
	}
	return s.describeCLI(service)
}

func (s *Session) describeCLI(service string) (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found and native gRPC reflection failed — run 'tinker deps' to install")
	}

	args := append(s.baseArgs(), s.Addr, "describe", service)
	out, err := exec.Command("grpcurl", args...).CombinedOutput() //nolint:gosec // args built from our own config
	return string(out), err
}

// Call invokes a gRPC method. Uses native client first, then CLI fallback.
func (s *Session) Call(method, data string) (string, error) {
	if s.Reflection {
		native := NewNativeClient(s.Addr, s.ProtoDir, s.Reflection)
		if out, err := native.Call(method, data); err == nil {
			return out, nil
		}
	}
	return s.callCLI(method, data)
}

func (s *Session) callCLI(method, data string) (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found and native gRPC call failed — run 'tinker deps' to install")
	}

	args := s.baseArgs()
	if data != "" {
		args = append(args, "-d", data)
	}
	args = append(args, s.Addr, method)

	out, err := exec.Command("grpcurl", args...).CombinedOutput() //nolint:gosec // args built from our own config
	return string(out), err
}

// Interactive opens an interactive gRPC REPL session.
// Uses evans if available, otherwise falls back to grpcurl.
func (s *Session) Interactive() error {
	if _, err := exec.LookPath("evans"); err == nil {
		return s.runEvans()
	}
	if _, err := exec.LookPath("grpcurl"); err == nil {
		return fmt.Errorf("evans not found (interactive REPL) — run 'tinker deps' to install\nUse 'tinker grpc call' with native gRPC instead")
	}
	return fmt.Errorf("no gRPC client found — run 'tinker deps' to install")
}

func (s *Session) runEvans() error {
	args := []string{}
	if s.ProtoDir != "" {
		args = append(args, "--path", s.ProtoDir)
	}
	if s.Reflection {
		args = append(args, "--reflection")
	}
	args = append(args, s.Addr)

	return runner.Interactive("evans", args...)
}
