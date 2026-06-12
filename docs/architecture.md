# Architecture

Tinker is a Go CLI tool built with the Cobra framework. This document describes the internal architecture, project structure, design principles, and how the various modules fit together.

## High-Level Architecture

```
┌──────────────────────────────────────────────────────┐
│                  tinker CLI (pure Go)                 │
│              cmd/tinker/*.go (Cobra commands)         │
├──────────────┬──────────────┬────────────────────────┤
│  DB Module   │  API Module  │  gRPC Module           │
│  (pgx/mysql  │  (native +   │  (native reflection    │
│   + CLI)     │   CLI)       │   + grpcurl/evans)     │
├──────────────┴──────────────┴────────────────────────┤
│  OpenAPI Parser │ Commands │ Envs │ Docker │ Plugins  │
├──────────────────────────────────────────────────────┤
│                   Config Layer                        │
│          (tinker.toml + .env resolver)                │
├──────────────────────────────────────────────────────┤
│              Auto-Detection Engine                    │
│           (tinker init scans project)                 │
└──────────────────────────────────────────────────────┘
```

## Project Structure

```
tinker/
├── cmd/tinker/               # CLI command layer (Cobra)
│   ├── main.go               # Root command, dashboard, version
│   ├── api.go                # API commands (GET/POST/PUT/DELETE, endpoints, explore, session)
│   ├── api_explorer.go       # Interactive API explorer REPL
│   ├── commands.go           # Custom command execution
│   ├── completion.go         # Shell completions (bash, zsh, fish, powershell)
│   ├── config.go             # Config show command
│   ├── dashboard.go          # Dashboard display helpers
│   ├── db.go                 # Database root + connect command
│   ├── db_extra.go           # db migrate, seed, explore (TUI)
│   ├── db_query.go           # db describe, indexes, schema, count, find, exec, tables
│   ├── deps.go               # Dependency checking/install
│   ├── docker.go             # Docker Compose inspection
│   ├── env.go                # Multi-environment commands
│   ├── grpc.go               # gRPC commands
│   ├── init.go               # Project initialization (auto-detect)
│   ├── log.go                # Log tailing and filtering
│   ├── make.go               # Makefile target runner
│   ├── plugin.go             # Plugin system
│   ├── run.go                # Run Go code
│   └── update.go             # Self-update
│
├── internal/                 # Internal packages (not importable outside this module)
│   ├── api/                  # HTTP client module
│   │   ├── spec.go           # OpenAPI spec parser (YAML/JSON)
│   │   ├── session.go        # HTTP session (request, auth, cookie jar)
│   │   ├── session_store.go  # Cookie/auth persistence to .tinker/session.json
│   │   ├── client.go         # HTTP client with gjson/jq response filtering
│   │   ├── interactive.go    # Interactive session logic
│   │   └── json_helpers.go   # JSON utilities
│   │
│   ├── config/               # Configuration system
│   │   ├── config.go         # Config types (Database, API, GRPC, Envs, etc.)
│   │   ├── load.go           # TOML + .env loading, env override application
│   │   ├── find.go           # Project root finder (walks up to find tinker.toml)
│   │   └── resolve.go        # Environment variable resolution + validation
│   │
│   ├── contract/             # tinker.toml template generation
│   │   └── contract.go       # Generates TOML from detection results
│   │
│   ├── db/                   # Database module
│   │   ├── session.go        # DB session (native + CLI, driver registration)
│   │   ├── query_schema.go   # Schema queries (native + CLI)
│   │   ├── query_tables.go   # Table listing
│   │   ├── query_describe.go # Table description
│   │   ├── query_indexes.go  # Index listing
│   │   ├── query_data.go     # Count, find, size queries
│   │   ├── query_exec.go     # SQL execution
│   │   ├── migrate.go        # Migration system (up, down, status, tracking)
│   │   ├── seed.go           # Seeding system (file splitting, statement execution)
│   │   ├── render.go         # go-pretty table rendering
│   │   ├── explore.go        # TUI database browser data layer
│   │   ├── explore_tui_*.go  # TUI model/update/view (Bubble Tea)
│   │   ├── usql.go           # CLI fallback methods
│   │   └── helpers.go        # Helper functions (quoteIdent, splitSQL, etc.)
│   │
│   ├── detect/               # Auto-detection engine
│   │   ├── detect.go         # Main detection orchestrator
│   │   ├── database.go       # Database detection (env vars, file scanning)
│   │   ├── api.go            # API detection (spec files, env vars)
│   │   ├── docker.go         # Docker Compose detection (YAML parsing, service heuristics)
│   │   ├── dirs.go           # Directory scanning helpers
│   │   ├── env.go            # .env parsing (no process mutation)
│   │   ├── grpc.go           # gRPC detection (proto dirs, env vars)
│   │   └── log.go            # Log file detection
│   │
│   ├── deps/                 # Dependency management
│   │   └── deps.go           # Tool checking and installation
│   │
│   ├── env/                  # Environment variable handling
│   │   └── env.go            # .env file parser + variable resolution
│   │
│   ├── grpc/                 # gRPC module
│   │   ├── session.go        # gRPC session configuration
│   │   ├── native.go         # Native gRPC client (server reflection, file descriptor parsing)
│   │   └── grpcurl.go        # grpcurl CLI fallback
│   │
│   ├── logfmt/               # Log formatting
│   │   └── logfmt.go         # Level detection, line formatting
│   │
│   ├── logtail/              # Log tailing
│   │   └── logtail.go        # Real-time log following
│   │
│   ├── make/                 # Makefile runner
│   │   └── make.go           # Target listing and execution
│   │
│   ├── plugin/               # Plugin system
│   │   └── plugin.go         # Registry, hooks, script loading, dispatch
│   │
│   ├── run/                  # Go code runner
│   │   └── run.go            # Parse and execute one-off Go code
│   │
│   ├── runner/               # Generic command runner
│   │   └── runner.go         # Shell command execution
│   │
│   └── ui/                   # Terminal UI
│       ├── styles.go         # Lipgloss styles, badges, banners, tables
│       └── highlight.go      # Chroma syntax highlighting (SQL, JSON, YAML, Go, TOML)
│
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── install.sh                # One-line install script
├── LICENSE                   # MIT license
└── README.md                 # Project documentation
```

## Design Principles

### 1. Native first, CLI fallback

Built-in pure Go drivers for PostgreSQL and MySQL handle one-shot queries without any external dependencies. External CLIs are only used for:
- **SQLite** queries (no pure Go driver in the binary due to size concerns)
- **Interactive sessions** (`db connect`, `grpc` REPL) which require REPL functionality that native drivers don't provide

This two-tier approach means most users get zero-dependency operation for daily tasks while still having access to full-featured interactive tools when needed.

### 2. Compose, don't reimplement

Tinker shells out to best-in-class open source tools for interactive sessions instead of reimplementing them:

| Purpose | Tool | Why we don't reimplement |
|---------|------|--------------------------|
| DB REPL | `pgcli`, `mycli`, `litecli`, `usql` | These tools have years of autocomplete, syntax highlighting, and UX polish |
| HTTP client | `curlie` | Full HTTP/2, redirect handling, and formatting |
| gRPC REPL | `evans` | Interactive proto field completion and streaming support |

### 3. Contract is declarative

The `tinker.toml` file is a simple, declarative configuration — not a Go interface or plugin API. This makes it language-agnostic: any project in any language can use Tinker by creating this file.

### 4. Auto-detect first, configure second

`tinker init` should work for 80% of cases. The auto-detection engine scans for common file patterns, environment variables, and directory structures. Manual configuration in `tinker.toml` is only needed for non-standard setups.

### 5. Cross-language by default

Tinker works with any project that has a `.env` file and a `tinker.toml`. It doesn't require Go code, Node.js, Python, or any specific language runtime. The only requirement is the Tinker binary itself.

## Key Dependencies

| Dependency | Purpose |
|-----------|---------|
| `spf13/cobra` | CLI framework (commands, flags, completions) |
| `pelletier/go-toml/v2` | TOML parsing for `tinker.toml` |
| `jackc/pgx/v5` | PostgreSQL native driver |
| `go-sql-driver/mysql` | MySQL native driver |
| `charmbracelet/bubbletea` | TUI framework for `db explore` |
| `charmbracelet/lipgloss` | Terminal styling (badges, colors, layouts) |
| `alecthomas/chroma/v2` | Syntax highlighting (SQL, JSON, YAML) |
| `jedib0t/go-pretty/v6` | Table rendering for query results |
| `tidwall/gjson` | JSON path filtering (native `--jq`) |
| `go.yaml.in/yaml/v3` | YAML parsing for OpenAPI specs |
| `google.golang.org/grpc` | Native gRPC client and server reflection |

## Configuration Loading Pipeline

```
1. Read tinker.toml
   ↓
2. Parse TOML into Config struct
   ↓
3. Load .env files into in-memory map (no process mutation)
   ↓
4. Apply environment overrides (if --env flag is set)
   ↓
5. Resolve env: references (replace "env:VAR" with actual values)
   ↓
6. Validate configuration
   ↓
7. Return resolved Config
```

Steps 3-5 happen in `config.LoadWithEnv()`. The `.env` parsing does not mutate `os.Environ()` — it creates an in-memory map that is used only for variable resolution. This is intentional: Tinker should not leak your `.env` values into the process environment.

## Session Management

### Database sessions

Each `tinker db` command creates a new `db.Session`, which:
1. Reads the database config and resolves the connection URL
2. Opens a native `database/sql` connection (for PostgreSQL/MySQL)
3. Verifies connectivity with `Ping()`
4. Defers `Close()` when the command finishes

Sessions are short-lived and not shared between commands. This ensures clean connection management and avoids state leakage.

### API sessions

API sessions support two modes:
- **One-shot** — Each `tinker api` call creates a session, makes the request, and exits
- **Persistent** — Cookies and auth state are saved to `.tinker/session.json` and restored on subsequent calls

The session store uses `net/http/cookiejar` for automatic cookie management and a JSON file for persistence. This enables login flows to work naturally across invocations.

### gRPC sessions

gRPC sessions are created per-command. The native client establishes a connection, performs the operation (list/describe), and closes the connection. No persistent state is maintained.

## Plugin System

The plugin system provides two extension mechanisms:

### Script plugins

Executable scripts placed in the `plugins/` directory are automatically loaded as command plugins:

```bash
plugins/
  migrate.sh     # becomes: tinker plugin run migrate
  seed.py        # becomes: tinker plugin run seed
  deploy.bash    # becomes: tinker plugin run deploy
```

Scripts are detected by their execute permission bit or known extensions (`.sh`, `.bash`, `.zsh`, `.py`, `.rb`, `.js`, `.ts`). The file extension is stripped for a cleaner alias.

### Hook system

The Go plugin API supports lifecycle hooks that plugins can register for:

| Hook | When it fires |
|------|--------------|
| `pre_db_query` | Before a database query |
| `post_db_query` | After a database query |
| `pre_api_request` | Before an API request |
| `post_api_request` | After an API response |
| `pre_grpc_call` | Before a gRPC call |
| `post_grpc_call` | After a gRPC call |
| `init` | During `tinker init` after auto-detection |

Hooks are dispatched in plugin registration order. If any handler returns an error, the chain stops and the error is propagated.

## Self-Update Mechanism

The `tinker update` command follows this flow:

1. Check for `git` and `go` binaries (required for building from source)
2. Fetch the latest version tag using a three-tier approach:
   - `git ls-remote --tags` (fast, no API rate limits)
   - GitHub Releases API (`/repos/{owner}/{repo}/releases/latest`)
   - GitHub Tags API (`/repos/{owner}/{repo}/tags?per_page=1`)
3. Clone the repository at the latest tag into a temp directory (shallow clone with `--depth 1`)
4. Build the binary with `go build` and install to `$HOME/go/bin/tinker`
5. If clone+build fails, fall back to `go install github.com/mvaliolahi/tinker/cmd/tinker@{tag}`

The version constant is set in `main.go` and can be overridden at build time with `-ldflags`:

```bash
go build -ldflags "-X main.version=$(git describe --tags)" ./cmd/tinker/
```
