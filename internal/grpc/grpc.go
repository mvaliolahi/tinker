// Package grpc provides gRPC service interaction capabilities by wrapping
// evans and grpcurl. It resolves connection details from tinker.toml
// and enables calling gRPC services from the terminal.
package grpc

import (
        "fmt"
        "os"
        "os/exec"

        "github.com/mvaliolahi/tinker/internal/config"
)

// Session represents a gRPC session configuration.
type Session struct {
        Addr       string
        ProtoDir   string
        Reflection bool
}

// NewSession creates a new gRPC session from config.
func NewSession(cfg *config.GRPCConfig) (*Session, error) {
        if cfg == nil {
                return nil, fmt.Errorf("no grpc configuration found in tinker.toml")
        }

        if cfg.ResolvedAddr == "" && cfg.ProtoDir == "" {
                return nil, fmt.Errorf("grpc addr or proto_dir must be configured")
        }

        return &Session{
                Addr:       cfg.ResolvedAddr,
                ProtoDir:   cfg.ProtoDir,
                Reflection: cfg.Reflection,
        }, nil
}

// Interactive opens an interactive gRPC REPL using evans if available,
// otherwise falls back to grpcurl.
func (s *Session) Interactive() error {
        // Try evans first (better REPL experience)
        if _, err := exec.LookPath("evans"); err == nil {
                return s.evansREPL()
        }

        // Fall back to grpcurl
        if _, err := exec.LookPath("grpcurl"); err == nil {
                return fmt.Errorf("evans is not installed (provides interactive REPL).\nInstall it: go install github.com/ktr0731/evans@latest\n\nYou can use 'tinker grpc call <service/method>' with grpcurl instead.\nInstall grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
        }

        return fmt.Errorf("no gRPC client found. Install one of:\n  evans:   go install github.com/ktr0731/evans@latest\n  grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
}

// Call invokes a specific gRPC method using grpcurl.
func (s *Session) Call(method string, data string) (string, error) {
        grpcurlPath, err := exec.LookPath("grpcurl")
        if err != nil {
                return "", fmt.Errorf("grpcurl is not installed. Install it:\n  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest")
        }

        args := []string{}

        if s.Reflection {
                args = append(args, "-plaintext")
        } else if s.ProtoDir != "" {
                args = append(args, "-import-path", s.ProtoDir)
                // Find proto files
                args = append(args, "-proto", fmt.Sprintf("%s/...", s.ProtoDir))
        }

        if data != "" {
                args = append(args, "-d", data)
        }

        args = append(args, s.Addr, method)

        cmd := exec.Command(grpcurlPath, args...)
        output, err := cmd.CombinedOutput()
        return string(output), err
}

// ListServices lists all available gRPC services.
func (s *Session) ListServices() (string, error) {
        grpcurlPath, err := exec.LookPath("grpcurl")
        if err != nil {
                return "", fmt.Errorf("grpcurl is not installed")
        }

        args := []string{}
        if !s.Reflection {
                if s.ProtoDir != "" {
                        args = append(args, "-import-path", s.ProtoDir)
                }
        }
        args = append(args, "-plaintext", s.Addr, "list")

        cmd := exec.Command(grpcurlPath, args...)
        output, err := cmd.CombinedOutput()
        return string(output), err
}

// DescribeService shows the details of a gRPC service.
func (s *Session) DescribeService(service string) (string, error) {
        grpcurlPath, err := exec.LookPath("grpcurl")
        if err != nil {
                return "", fmt.Errorf("grpcurl is not installed")
        }

        args := []string{}
        if !s.Reflection {
                if s.ProtoDir != "" {
                        args = append(args, "-import-path", s.ProtoDir)
                }
        }
        args = append(args, "-plaintext", s.Addr, "describe", service)

        cmd := exec.Command(grpcurlPath, args...)
        output, err := cmd.CombinedOutput()
        return string(output), err
}

// evansREPL opens an interactive evans REPL.
func (s *Session) evansREPL() error {
        args := []string{}

        if s.ProtoDir != "" {
                args = append(args, "--path", s.ProtoDir)
        }

        if s.Reflection {
                args = append(args, "--reflection")
        }

        args = append(args, s.Addr)

        cmd := exec.Command("evans", args...)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        return cmd.Run()
}
