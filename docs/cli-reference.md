# CLI Reference

Complete reference for all Tinker commands, flags, and aliases.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project directory (default: current directory) |
| `--env` | `-e` | Environment name (e.g., `staging`, `production`) |

These flags are persistent — they can be used with any subcommand.

## Root Command

```bash
tinker
```

When run without a subcommand, Tinker displays the project dashboard — a visual overview of detected services including database, API, gRPC, Docker Compose, and missing dependencies.

## Database Commands

### `tinker db`

Without a subcommand, opens an interactive database session (same as `tinker db connect`).

| Command | Alias | Args | Description |
|---------|-------|------|-------------|
| `tinker db tables` | `ls` | — | List all tables |
| `tinker db describe <table>` | `desc` | table name | Show column schema |
| `tinker db indexes <table>` | `idx` | table name | Show table indexes |
| `tinker db schema <table>` | `s` | table name | Show `CREATE TABLE` statement |
| `tinker db count <table> [where]` | `c` | table, optional WHERE | Count rows |
| `tinker db find <table> <id>` | `f` | table, id value | Find row by ID |
| `tinker db exec "<sql>"` | `e`, `sql` | SQL statement | Execute arbitrary SQL |
| `tinker db ping` | — | — | Test database connectivity |
| `tinker db size` | — | — | Row counts for all tables |
| `tinker db connect` | — | — | Open interactive DB session |
| `tinker db explore` | — | — | Full-screen TUI database browser |
| `tinker db seed [path]` | — | optional path | Run seed files |
| `tinker db migrate up` | — | — | Run pending migrations |
| `tinker db migrate down` | `rollback` | — | Rollback last migration |
| `tinker db migrate status` | `st` | — | Show migration status |

## API Commands

### `tinker api [method] [path] [body]`

Call an HTTP endpoint directly. Method defaults to `GET` if only a path is provided.

| Flag | Short | Description |
|------|-------|-------------|
| `--jq` | `-q` | Filter response with gjson path or jq expression |

| Command | Description |
|---------|-------------|
| `tinker api GET /users` | GET request |
| `tinker api POST /users '{"name":"Ali"}'` | POST with body |
| `tinker api PUT /users/1 '{"name":"Updated"}'` | PUT with body |
| `tinker api DELETE /users/1` | DELETE request |

### `tinker api endpoints`

| Alias | `ep` |
|-------|------|

| Flag | Short | Description |
|------|-------|-------------|
| `--tag` | `-t` | Filter endpoints by OpenAPI tag |

### `tinker api explore`

Opens an interactive API explorer REPL. Requires an OpenAPI spec configured in `tinker.toml`.

### `tinker api session`

| Command | Description |
|---------|-------------|
| `tinker api session show` | View current session state (cookies, auth, headers) |
| `tinker api session clear` | Clear all persisted session data |

## gRPC Commands

### `tinker grpc`

Without a subcommand, opens an interactive gRPC session (requires `evans`).

| Command | Args | Description |
|---------|------|-------------|
| `tinker grpc list` | — | List gRPC services (native, no grpcurl needed) |
| `tinker grpc describe <service>` | service name | Describe a service (native) |
| `tinker grpc call <method> [data]` | method, optional JSON | Call a gRPC method (requires grpcurl) |

## Log Commands

### `tinker log [path]`

Display log file contents. Uses files from `[log]` config, scans for logs, or accepts a manual path.

| Flag | Short | Description |
|------|-------|-------------|
| `--tail` | `-n` | Show last N lines (0 = all) |
| `--level` | `-l` | Filter by log level (`error`, `warn`, `info`, `debug`) |
| `--grep` | `-g` | Filter by text pattern (case-insensitive) |

| Subcommand | Description |
|------------|-------------|
| `tinker log tail [path]` | Follow (tail) a log file in real-time |
| `tinker log list` | List configured log files |

## Environment Commands

### `tinker env`

| Command | Alias | Description |
|---------|-------|-------------|
| `tinker env list` | `ls` | List available environments |
| `tinker env show <name>` | — | Show configuration for a specific environment |

## Custom Commands

### `tinker cmd [name] [args...]`

Run a custom project command defined in the `[commands]` section of `tinker.toml`. Extra arguments are appended to the command string.

| Command | Alias | Description |
|---------|-------|-------------|
| `tinker cmd <name>` | — | Run a custom command |
| `tinker cmd list` | `ls` | List available custom commands |

Commands are executed via your system shell, so pipes, redirects, and shell features work natively.

## Plugin Commands

### `tinker plugin`

| Command | Alias | Args | Description |
|---------|-------|------|-------------|
| `tinker plugin list` | `ls` | — | List loaded plugins |
| `tinker plugin run <name> [args...]` | — | plugin name, optional args | Run a script plugin |

## Docker Commands

### `tinker docker`

| Command | Alias | Description |
|---------|-------|-------------|
| `tinker docker list` | `ls` | List Docker Compose services with detected types |
| `tinker docker info` | — | Show detailed service information |

Without a subcommand, shows a summary of Docker Compose services.

## Other Commands

### `tinker init`

Scan your project directory and generate a `tinker.toml` configuration file. Automatically detects databases, APIs, gRPC services, log files, and Docker Compose services.

### `tinker config show`

Display the resolved configuration, including environment variable substitution and environment overrides.

### `tinker deps`

| Command | Description |
|---------|-------------|
| `tinker deps list` | List dependency status (installed vs. missing) |
| `tinker deps install` | Install missing dependencies |

### `tinker make [target] [args...]`

Run a Makefile target. Extra arguments are passed to `make`.

| Command | Description |
|---------|-------------|
| `tinker make list` | List available Makefile targets |

### `tinker run [code]`

Execute one-off Go code in the project context. Useful for quick computations or data transformations.

```bash
tinker run 'fmt.Println("Hello from Tinker!")'
```

### `tinker update`

Update Tinker to the latest version. Fetches the latest git tag from the repository, clones the source at that tag, builds the binary, and installs it to `$HOME/go/bin/tinker`. Falls back to `go install` if the clone+build approach fails.

### `tinker version`

Print the current Tinker version.

### `tinker completion [shell]`

Generate shell completion scripts. Supports `bash`, `zsh`, `fish`, and `powershell`.

```bash
# Bash
tinker completion bash > ~/.tinker-completion.bash
echo 'source ~/.tinker-completion.bash' >> ~/.bashrc

# Zsh
tinker completion zsh > "${fpath[1]}/_tinker"

# Fish
tinker completion fish > ~/.config/fish/completions/tinker.fish
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (command failed, configuration invalid, etc.) |

Tinker suppresses usage messages on errors (`SilenceErrors: true`) and displays errors with styled formatting via the `ui.Error()` function.
