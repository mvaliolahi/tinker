package contract

import (
        "os"
        "path/filepath"
        "strings"
        "testing"

        "github.com/mvaliolahi/tinker/internal/detect"
)

// ---------------------------------------------------------------------------
// findSection
// ---------------------------------------------------------------------------

func TestFindSection(t *testing.T) {
        tests := []struct {
                name    string
                content string
                section string
                want    int // -1 means not found
        }{
                {
                        name:    "simple section at start",
                        content: "[database]\nsource = \"env:DB\"\n",
                        section: "database",
                        want:    0,
                },
                {
                        name:    "section in middle",
                        content: "[database]\nsource = \"x\"\n\n[api]\nbase_url = \"y\"\n",
                        section: "api",
                        want:    strings.Index("[database]\nsource = \"x\"\n\n[api]\nbase_url = \"y\"\n", "[api]"),
                },
                {
                        name:    "section not found",
                        content: "[database]\nsource = \"x\"\n",
                        section: "grpc",
                        want:    -1,
                },
                {
                        name:    "skip table array [[log]]",
                        content: "[[log]]\nfile = \"a\"\n\n[log]\nfiles = []\n",
                        section: "log",
                        want:    18, // index in truncated content after skipping [[log]]
                },
                {
                        name:    "empty content",
                        content: "",
                        section: "database",
                        want:    -1,
                },
                {
                        name:    "section not at start of line",
                        content: "# comment [database]\n[database]\nsource = \"x\"\n",
                        section: "database",
                        want:    10, // index in truncated content after skipping comment
                },
        }
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := findSection(tt.content, tt.section)
                        if got != tt.want {
                                t.Errorf("findSection() = %d, want %d", got, tt.want)
                        }
                })
        }
}

// ---------------------------------------------------------------------------
// findNextSection
// ---------------------------------------------------------------------------

func TestFindNextSection(t *testing.T) {
        tests := []struct {
                name    string
                content string
                after   int
                want    int
        }{
                {
                        name:    "next section exists",
                        content: "[database]\nsource = \"x\"\n\n[api]\nbase_url = \"y\"\n",
                        after:   0, // at [database]
                        want:    25, // position of [api] in content
                },
                {
                        name:    "no next section — returns len",
                        content: "[database]\nsource = \"x\"\n",
                        after:   0,
                        want:    len("[database]\nsource = \"x\"\n"),
                },
                {
                        name:    "next section at EOF boundary",
                        content: "[log]\nfiles = []\n\n[commands]\nmigrate = \"x\"\n",
                        after:   0,
                        want:    18, // position of [commands] in content
                },
        }
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := findNextSection(tt.content, tt.after)
                        if got != tt.want {
                                t.Errorf("findNextSection() = %d, want %d", got, tt.want)
                        }
                })
        }
}

// ---------------------------------------------------------------------------
// Generate
// ---------------------------------------------------------------------------

func TestGenerate_FullResult(t *testing.T) {
        dir := t.TempDir()

        result := &detect.Result{
                Database: &detect.DatabaseResult{
                        Source:     "env:DATABASE_URL",
                        Type:       "postgres",
                        MigrateDir: "db/migrate",
                        SeedDir:    "db/seed",
                },
                API: &detect.APIResult{
                        BaseURL:  "https://api.example.com",
                        Spec:     "openapi.yaml",
                        Auth:     "env:API_TOKEN",
                        AuthType: "bearer",
                },
                GRPC: &detect.GRPCResult{
                        Addr:       "localhost:9090",
                        ProtoDir:   "proto",
                        Reflection: true,
                },
                Log: &detect.LogResult{
                        Files: []string{"logs/app.log", "error.log"},
                },
                Docker: &detect.DockerResult{
                        ComposeFile: "docker-compose.yml",
                },
        }

        if err := Generate(result, dir); err != nil {
                t.Fatalf("Generate() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatalf("reading tinker.toml: %v", err)
        }
        content := string(data)

        // Verify header
        if !strings.Contains(content, "# Tinker Configuration") {
                t.Error("expected header comment in tinker.toml")
        }

        // Verify database section
        if !strings.Contains(content, "[database]") {
                t.Error("expected [database] section")
        }
        if !strings.Contains(content, `source = "env:DATABASE_URL"`) {
                t.Error("expected database source")
        }
        if !strings.Contains(content, `type = "postgres"`) {
                t.Error("expected database type")
        }
        if !strings.Contains(content, `migrate_dir = "db/migrate"`) {
                t.Error("expected migrate_dir")
        }
        if !strings.Contains(content, `seed_dir = "db/seed"`) {
                t.Error("expected seed_dir")
        }

        // Verify API section
        if !strings.Contains(content, "[api]") {
                t.Error("expected [api] section")
        }
        if !strings.Contains(content, `base_url = "https://api.example.com"`) {
                t.Error("expected api base_url")
        }
        if !strings.Contains(content, `spec = "openapi.yaml"`) {
                t.Error("expected api spec")
        }
        if !strings.Contains(content, `auth = "env:API_TOKEN"`) {
                t.Error("expected api auth")
        }
        if !strings.Contains(content, `auth_type = "bearer"`) {
                t.Error("expected api auth_type")
        }

        // Verify gRPC section
        if !strings.Contains(content, "[grpc]") {
                t.Error("expected [grpc] section")
        }
        if !strings.Contains(content, `addr = "localhost:9090"`) {
                t.Error("expected grpc addr")
        }
        if !strings.Contains(content, `proto_dir = "proto"`) {
                t.Error("expected grpc proto_dir")
        }
        if !strings.Contains(content, "reflection = true") {
                t.Error("expected grpc reflection")
        }

        // Verify log section
        if !strings.Contains(content, "[log]") {
                t.Error("expected [log] section")
        }
        if !strings.Contains(content, `"logs/app.log"`) {
                t.Error("expected log file logs/app.log")
        }
        if !strings.Contains(content, `"error.log"`) {
                t.Error("expected log file error.log")
        }

        // Verify Docker comment
        if !strings.Contains(content, "Docker Compose detected: docker-compose.yml") {
                t.Error("expected Docker comment")
        }
}

func TestGenerate_MinimalResult(t *testing.T) {
        dir := t.TempDir()

        result := &detect.Result{
                Database: &detect.DatabaseResult{
                        Source: "env:DB_PATH",
                        Type:   "sqlite3",
                },
        }

        if err := Generate(result, dir); err != nil {
                t.Fatalf("Generate() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatalf("reading tinker.toml: %v", err)
        }
        content := string(data)

        if !strings.Contains(content, "[database]") {
                t.Error("expected [database] section")
        }
        if strings.Contains(content, "[api]") {
                t.Error("did not expect [api] section for nil API")
        }
        if strings.Contains(content, "[grpc]") {
                t.Error("did not expect [grpc] section for nil GRPC")
        }
        if strings.Contains(content, "[log]") {
                t.Error("did not expect [log] section for nil Log")
        }
}

func TestGenerate_EmptyResult(t *testing.T) {
        dir := t.TempDir()

        result := &detect.Result{}

        if err := Generate(result, dir); err != nil {
                t.Fatalf("Generate() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatalf("reading tinker.toml: %v", err)
        }
        content := string(data)

        // Header and commands comment should still exist
        if !strings.Contains(content, "# Tinker Configuration") {
                t.Error("expected header comment")
        }
        if !strings.Contains(content, "[commands]") {
                t.Error("expected [commands] comment section")
        }
        // No data sections
        if strings.Contains(content, "[database]") {
                t.Error("did not expect [database] section for nil result fields")
        }
}

// ---------------------------------------------------------------------------
// SaveLogConfig
// ---------------------------------------------------------------------------

func TestSaveLogConfig_ReplaceExisting(t *testing.T) {
        dir := t.TempDir()

        // Create initial tinker.toml with a [log] section
        initial := `[database]
source = "env:DB_URL"
type = "postgres"

[log]
files = [
  "old.log",
]

[api]
base_url = "http://localhost"
`
        if err := os.WriteFile(filepath.Join(dir, "tinker.toml"), []byte(initial), 0644); err != nil {
                t.Fatal(err)
        }

        newFiles := []string{"app.log", "error.log"}
        if err := SaveLogConfig(newFiles, dir); err != nil {
                t.Fatalf("SaveLogConfig() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatal(err)
        }
        content := string(data)

        // Should contain new log files
        if !strings.Contains(content, `"app.log"`) {
                t.Error("expected app.log in updated config")
        }
        if !strings.Contains(content, `"error.log"`) {
                t.Error("expected error.log in updated config")
        }
        // Should not contain old log file
        if strings.Contains(content, `"old.log"`) {
                t.Error("did not expect old.log in updated config")
        }
        // Other sections should remain
        if !strings.Contains(content, `[database]`) {
                t.Error("expected [database] section to remain")
        }
        if !strings.Contains(content, `[api]`) {
                t.Error("expected [api] section to remain")
        }
}

func TestSaveLogConfig_InsertBeforeCommands(t *testing.T) {
        dir := t.TempDir()

        // Create tinker.toml without [log] but with [commands] comment
        initial := `[database]
source = "env:DB_URL"
type = "postgres"

# [commands]
# migrate = "go run ./cmd/migrate"
`
        if err := os.WriteFile(filepath.Join(dir, "tinker.toml"), []byte(initial), 0644); err != nil {
                t.Fatal(err)
        }

        newFiles := []string{"app.log"}
        if err := SaveLogConfig(newFiles, dir); err != nil {
                t.Fatalf("SaveLogConfig() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatal(err)
        }
        content := string(data)

        if !strings.Contains(content, "[log]") {
                t.Error("expected [log] section to be inserted")
        }
        if !strings.Contains(content, `"app.log"`) {
                t.Error("expected app.log in inserted config")
        }
}

func TestSaveLogConfig_AppendAtEnd(t *testing.T) {
        dir := t.TempDir()

        // Create minimal tinker.toml with no [log] or [commands]
        initial := `[database]
source = "env:DB_URL"
type = "sqlite3"
`
        if err := os.WriteFile(filepath.Join(dir, "tinker.toml"), []byte(initial), 0644); err != nil {
                t.Fatal(err)
        }

        newFiles := []string{"dev.log"}
        if err := SaveLogConfig(newFiles, dir); err != nil {
                t.Fatalf("SaveLogConfig() error: %v", err)
        }

        data, err := os.ReadFile(filepath.Join(dir, "tinker.toml"))
        if err != nil {
                t.Fatal(err)
        }
        content := string(data)

        if !strings.Contains(content, "[log]") {
                t.Error("expected [log] section to be appended")
        }
        if !strings.Contains(content, `"dev.log"`) {
                t.Error("expected dev.log in appended config")
        }
}

func TestSaveLogConfig_NoTOMLFile(t *testing.T) {
        dir := t.TempDir()

        err := SaveLogConfig([]string{"app.log"}, dir)
        if err == nil {
                t.Error("expected error when tinker.toml does not exist")
        }
}
