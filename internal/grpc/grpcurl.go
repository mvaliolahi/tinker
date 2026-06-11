package grpc

import (
	"fmt"
	"os"
	"os/exec"
)

func (s *Session) baseArgs() []string {
	args := []string{"-plaintext"}
	if !s.Reflection && s.ProtoDir != "" {
		args = append(args, "-import-path", s.ProtoDir, "-proto", s.ProtoDir+"/...")
	}
	return args
}

func (s *Session) ListServices() (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found — install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
	}

	args := append(s.baseArgs(), s.Addr, "list")
	out, err := exec.Command("grpcurl", args...).CombinedOutput()
	return string(out), err
}

func (s *Session) Describe(service string) (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found")
	}

	args := append(s.baseArgs(), s.Addr, "describe", service)
	out, err := exec.Command("grpcurl", args...).CombinedOutput()
	return string(out), err
}

func (s *Session) Call(method, data string) (string, error) {
	if _, err := exec.LookPath("grpcurl"); err != nil {
		return "", fmt.Errorf("grpcurl not found")
	}

	args := s.baseArgs()
	if data != "" {
		args = append(args, "-d", data)
	}
	args = append(args, s.Addr, method)

	out, err := exec.Command("grpcurl", args...).CombinedOutput()
	return string(out), err
}

func (s *Session) Interactive() error {
	if _, err := exec.LookPath("evans"); err == nil {
		return s.runEvans()
	}
	if _, err := exec.LookPath("grpcurl"); err == nil {
		return fmt.Errorf("evans not found (interactive REPL) — install: go install github.com/ktr0731/evans@latest\nUse 'tinker grpc call' with grpcurl instead")
	}
	return fmt.Errorf("no gRPC client found — install evans or grpcurl")
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

	cmd := exec.Command("evans", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}
