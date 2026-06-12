# Contributing

Contributions to Tinker are welcome! This guide covers the development setup, coding standards, and contribution workflow.

## Development Setup

### Prerequisites

- **Go 1.25+** — Tinker uses the latest Go features and module system
- **Git** — For version control
- **golangci-lint** — For code linting (`go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest`)

### Clone and build

```bash
git clone https://github.com/mvaliolahi/tinker.git
cd tinker
go build ./cmd/tinker
```

The built binary will be at `./tinker`. You can install it to your GOPATH:

```bash
go install ./cmd/tinker
```

### Run tests

```bash
go test ./...
```

### Run linter

```bash
golangci-lint run ./...
```

Tinker uses `golangci-lint` with the default configuration. All linting errors should be resolved before submitting a PR.

## Project Structure

See [Architecture](architecture.md) for the complete project structure. Key points for contributors:

- **`cmd/tinker/`** — All Cobra command definitions. Each file corresponds to a top-level command or command group. Commands should be thin — they parse arguments, create sessions, and delegate to internal packages.
- **`internal/`** — All business logic. Packages are organized by domain (db, api, grpc, config, etc.). Code in `internal/` cannot be imported by external projects.
- **`internal/ui/`** — All terminal output. UI functions should be used consistently across commands for a uniform look and feel.

## Coding Standards

### Command structure

Follow the existing pattern for new commands:

1. Create a new file in `cmd/tinker/` named after the command (e.g., `feature.go`)
2. Define a `featureCmd()` function that returns a `*cobra.Command`
3. Add subcommands as needed using `cmd.AddCommand()`
4. Register the command in `main.go`'s `root.AddCommand(...)` call
5. Use `RunE` instead of `Run` for error-returning command handlers
6. Use `loadConfig()` to get the resolved configuration
7. Use `newDBSession()`, `newAPISession()`, etc. to create domain sessions
8. Always `defer session.Close()` for database sessions

### UI consistency

Use the `ui` package for all terminal output:

- `ui.Bold()`, `ui.Dim()`, `ui.Accent()` — Text styling
- `ui.DBLabel()`, `ui.APILabel()`, `ui.GRPCLabel()` — Colored badges
- `ui.KeyValue()` — Labeled key-value pairs
- `ui.Bullet()` — Bullet points with labels
- `ui.Success()`, `ui.Error()`, `ui.Warning()` — Status indicators
- `ui.Hint()` — User-facing hints and suggestions
- `ui.Step()`, `ui.StepDone()` — Progress steps
- `ui.Table()` — Aligned tables
- `ui.Banner()` — Version banner

### Error handling

- Return errors from `RunE` rather than calling `os.Exit()` directly (except in migration/seed commands where exit codes are used for CI pipelines)
- Use `fmt.Errorf("context: %w", err)` for error wrapping to preserve the error chain
- Display user-friendly error messages with `ui.Error()` for validation errors
- Let Cobra handle usage display for argument errors

### Native first, CLI fallback

When adding database functionality:

1. Implement the native `database/sql` version first (in `internal/db/`)
2. Add a CLI fallback that spawns an external tool
3. In the command handler, try native first and fall back to CLI:

```go
if s.HasNativeConn() {
    out, err := s.NativeMethod()
    if err == nil {
        fmt.Print(out)
        return nil
    }
    // Fall through to CLI fallback
}
out, err := s.CLIMethod()
fmt.Print(out)
return err
```

### Configuration

When adding new configuration options:

1. Add the field to the appropriate struct in `internal/config/config.go`
2. Add a TOML tag for serialization
3. Add env variable resolution in `internal/config/resolve.go` if the field supports `env:` prefix
4. Add auto-detection logic in `internal/detect/` if the field can be auto-detected
5. Update the contract generator in `internal/contract/contract.go`
6. Update the `Validate()` method if the field has validation rules

## Contribution Workflow

1. **Fork** the repository on GitHub
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/my-feature
   ```
3. **Make your changes** following the coding standards above
4. **Add tests** for new functionality where applicable
5. **Ensure linting passes**:
   ```bash
   golangci-lint run ./...
   ```
6. **Ensure tests pass**:
   ```bash
   go test ./...
   ```
7. **Commit with a descriptive message** following conventional commits:
   - `feat: add new database command`
   - `fix: resolve connection timeout for PostgreSQL`
   - `docs: update API documentation`
   - `refactor: simplify config resolution`
8. **Push** your branch and open a Pull Request against `main`

## Pull Request Guidelines

- **Small, focused PRs** — Each PR should address one concern
- **Descriptive title** — Use conventional commit format in the PR title
- **Explain the "why"** — The PR description should explain the motivation, not just what was changed
- **Update documentation** — If you add a feature, update the relevant docs in `docs/`
- **Backward compatibility** — Avoid breaking changes to the CLI interface or `tinker.toml` format. If a breaking change is necessary, discuss it in an issue first

## Reporting Issues

When reporting bugs, please include:

1. Tinker version (`tinker version`)
2. Operating system and architecture
3. Steps to reproduce
4. Expected behavior
5. Actual behavior (including error messages)
6. Your `tinker.toml` (redacted of sensitive values)
7. Your `.env` (redacted of sensitive values)

## License

By contributing to Tinker, you agree that your contributions will be licensed under the [MIT License](LICENSE).
