package grpc

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/config"
)

type Session struct {
	Addr       string
	ProtoDir   string
	Reflection bool
}

func NewSession(cfg *config.GRPC) (*Session, error) {
	if cfg == nil {
		return nil, fmt.Errorf("no [grpc] section in tinker.toml")
	}
	if cfg.ResolvedAddr == "" && cfg.ProtoDir == "" {
		return nil, fmt.Errorf("grpc addr or proto_dir must be set")
	}

	return &Session{
		Addr:       cfg.ResolvedAddr,
		ProtoDir:   cfg.ProtoDir,
		Reflection: cfg.Reflection,
	}, nil
}
