package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var addrEnvNames = []string{"GRPC_ADDR", "GRPC_ADDRESS", "GRPC_HOST"}
var protoDirs = []string{"proto", "protos", "api/proto", "api/protos", "pkg/proto"}

func (d *Detector) detectGRPC() *GRPCResult {
	r := &GRPCResult{}

	for _, name := range addrEnvNames {
		if d.getEnv(name) != "" {
			r.Addr = fmt.Sprintf("env:%s", name)
			break
		}
	}

	for _, dir := range protoDirs {
		full := filepath.Join(d.dir, dir)
		if info, err := os.Stat(full); err == nil && info.IsDir() && hasProtoFiles(full) {
			r.ProtoDir = dir
			break
		}
	}

	if r.Addr == "" && r.ProtoDir == "" {
		return nil
	}

	r.Reflection = true
	return r
}

func hasProtoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".proto") {
			return true
		}
	}
	return false
}
