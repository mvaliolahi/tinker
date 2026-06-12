<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/v/tag/mvaliolahi/tinker?style=flat-square&label=version" alt="Version">
  <img src="https://img.shields.io/github/actions/workflow/status/mvaliolahi/tinker/ci.yml?style=flat-square&label=CI" alt="CI">
</p>

<h1 align="center">Tinker</h1>

<p align="center">
  <strong>A project-aware CLI for database, API &amp; gRPC interaction</strong><br>
  <em>Inspired by Laravel Tinker — built for Go, works with any language</em>
</p>

<p align="center">
  <a href="#installation">Install</a> &middot;
  <a href="#quick-start">Quick Start</a> &middot;
  <a href="#configuration">Configuration</a> &middot;
  <a href="#commands">Commands</a> &middot;
  <a href="#contributing">Contribute</a>
</p>

---

Tinker comes with **built-in database drivers** (PostgreSQL, MySQL) for zero-dependency one-shot queries, and composes best-in-class open source tools (`usql`, `curlie`, `grpcurl`/`evans`) for interactive sessions. It reads your project's `tinker.toml` and `.env` to automatically configure connections — so you can jump straight into working with your project.

```bash
cd your-project
tinker init      # scans .env, specs, protos, compose files → tinker.toml
tinker db tables  # works instantly — no CLI tools needed for Postgres/MySQL
tinker api GET /users
tinker grpc list
```

## Why Tinker?

If you've used Laravel Tinker, you know the workflow: open a shell, query the database, call an API endpoint, test a service — all within your project context. Go doesn't have an equivalent because it's compiled, not interpreted.

**Tinker takes a different approach.** Instead of trying to be a REPL for Go code, it:

1. **Reads your project's existing config** (`.env`, `tinker.toml`) — no wrapper abstractions
2. **Composes proven OSS tools** — `usql` for DB, `curlie` for HTTP, `evans`/`grpcurl` for gRPC
3. **Defines a simple contract** — any project that implements `tinker.toml` gets the full experience
4. **Works with any language** — not just Go! Any project with a `.env` and `tinker.toml` works

## Installation

### One-line install (recommended)

Installs the binary and configures your PATH:

```bash
curl -fsSL https://raw.githubusercontent.com/mvaliolahi/tinker/main/install.sh | bash
```

### Install with Go

```bash
go install github.com/mvaliolahi/tinker/cmd/tinker@latest
```

Make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH`.

### Homebrew

```bash
brew tap mvaliolahi/tap
brew install tinker
```

## Quick Start

### 1. Initialize

```bash
cd your-project
tinker init
```

Tinker scans your project and generates a `tinker.toml` by detecting:

- `.env` files with `DATABASE_URL`, `DB_PATH`, `API_BASE_URL`, `GRPC_ADDR`, etc.
- OpenAPI / Swagger spec files (`openapi.yaml`, `swagger.json`, …)
- `.proto` directories
- SQLite database files (`.db`, `.sqlite`, `.sqlite3`)
- Docker Compose files

### 2. Run

```bash
tinker db tables       # list tables — works out of the box for Postgres/MySQL
tinker db explore      # full-screen TUI database browser
tinker api endpoints   # list API endpoints from your OpenAPI spec
tinker api GET /users  # call an endpoint
```

That's it. No manual configuration needed for most projects.

## Commands

### Database

| Command | Alias | Description |
|---------|-------|-------------|
| `tinker db tables` | `ls` | List all tables |
| `tinker db describe <table>` | `desc` | Show column schema |
| `tinker db indexes <table>` | `idx` | Show table indexes |
| `tinker db schema <table>` | `s` | Show `CREATE TABLE` statement |
| `tinker db count <table> [where]` | `c` | Count rows (optional `WHERE`) |
| `tinker db find <table> <id>` | `f` | Find a row by ID |
| `tinker db exec "<sql>"` | `e`, `sql` | Run arbitrary SQL |
| `tinker db ping` | | Test database connectivity |
| `tinker db size` | | Show row counts for all tables |
| `tinker db connect` | | Open interactive DB session (`usql`/`pgcli`/`mycli`/`litecli`) |
| `tinker db explore` | | Full-screen TUI database browser |

#### Migrations

Tinker includes a built-in migration system that tracks applied versions in a `_tinker_migrations` table.

```bash
tinker db migrate up      # run pending migrations
tinker db migrate down    # rollback the last migration
tinker db migrate status  # show applied vs. pending
```

Migration files follow the `NNN_description.up.sql` / `NNN_description.down.sql` convention:

```
migrations/
  001_create_users.up.sql
  001_create_users.down.sql
  002_create_posts.up.sql
  002_create_posts.down.sql
```

The migrations directory is auto-detected (`migrations/`, `db/migrations/`, `backend/migrations/`, …) or configured explicitly:

```toml
[database]
migrate_dir = "db/migrations"
```

#### Seeding

```bash
tinker db seed                # run all .sql files in seed/ directory
tinker db seed seed/users.sql # run a specific seed file
tinker db seed fixtures/      # run all .sql files in a directory
```

Seed files are split by semicolons (respecting quoted strings and `--` comments) and executed statement by statement. The seed directory is auto-detected or configured:

```toml
[database]
seed_dir = "db/seed"
```

### API

```bash
tinker api endpoints                # list all endpoints (alias: ep)
tinker api endpoints --tag users    # filter by tag
tinker api explore                  # interactive API explorer

# Call endpoints directly
tinker api GET /users
tinker api POST /users '{"name": "Ali"}'
tinker api PUT /users/1 '{"name": "Updated"}'
tinker api DELETE /users/1

# Filter responses
tinker api GET /users -q "data.0.name"            # gjson path (no jq needed)
tinker api GET /users -q '.[] | select(.active)'  # complex jq (falls back to jq CLI)
```

### gRPC

```bash
tinker grpc                           # Interactive REPL (evans)
tinker grpc list                      # List services
tinker grpc describe UserService       # Describe a service
tinker grpc call UserService/GetUser '{"id": 1}'
```

### Custom Commands

Define project-specific commands in `tinker.toml`:

```toml
[commands]
migrate = "go run ./cmd/migrate"
seed    = "go run ./cmd/seed"
test    = "go test ./..."
lint    = "golangci-lint run"
dev     = "air"
build   = "go build -o bin/app ./cmd/app"
```

```bash
tinker cmd migrate          # run a custom command
tinker cmd test -v -run X  # append extra arguments
tinker cmd list             # list available commands
```

Commands are executed via your system shell, so pipes, redirects, and shell features work out of the box.

### Multi-Environment

```bash
tinker --env staging db ping       # use staging configuration
tinker --env production db tables  # use production configuration
tinker env list                    # list available environments
tinker env show staging            # show staging overrides
```

### Other Commands

```bash
tinker              # dashboard — project overview with detected services
tinker config show  # display resolved configuration
tinker docker list  # list Docker Compose services
tinker deps list    # check which companion tools are installed
tinker deps install # install missing tools
tinker make build   # run a Makefile target
tinker make list    # list available Makefile targets
tinker version      # print version
tinker completion   # generate shell completions (bash, zsh, fish, powershell)
```

## Configuration

### `tinker.toml` — Full Spec

The `tinker.toml` file is Tinker's contract with your project. It's declarative, minimal, and language-agnostic.

```toml
# ── Database ──────────────────────────────────────────────
[database]
# Connection string source:
#   "env:VAR_NAME"  — read from environment variable (supports .env files)
#   "postgres://…"  — direct DSN
source = "env:DATABASE_URL"
# Database type: postgres | mysql | sqlite3
type = "postgres"
# Override the driver name (optional)
driver = ""
# Migration directory (relative to project root, auto-detected if not set)
migrate_dir = "migrations"
# Seed directory (relative to project root, auto-detected if not set)
seed_dir = "seed"

# ── API ───────────────────────────────────────────────────
[api]
base_url  = "env:API_BASE_URL"
spec      = "openapi.yaml"      # OpenAPI/Swagger spec (optional)
auth      = "env:API_TOKEN"     # Auth token source
auth_type = "bearer"            # bearer | basic | api_key | raw

[api.headers]                    # Additional default headers
X-Custom-Header = "value"

# ── gRPC ──────────────────────────────────────────────────
[grpc]
addr       = "env:GRPC_ADDR"
proto_dir  = "./proto"
reflection = true

# ── Custom Commands ───────────────────────────────────────
[commands]
migrate = "go run ./cmd/migrate"
seed    = "go run ./cmd/seed"
test    = "go test ./..."

# ── Multi-Environment ─────────────────────────────────────
[envs.staging.database]
source = "env:STAGING_DATABASE_URL"

[envs.staging.api]
base_url = "env:STAGING_API_BASE_URL"
auth     = "env:STAGING_API_TOKEN"

[envs.production.database]
source = "env:PRODUCTION_DATABASE_URL"

[envs.production.api]
base_url = "env:PRODUCTION_API_BASE_URL"
auth     = "env:PRODUCTION_API_TOKEN"
```

### Environment Variable Resolution

Any value starting with `env:` is resolved from the environment. Tinker automatically loads `.env` files, so this works:

```bash
# .env
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
API_BASE_URL=http://localhost:8080
API_TOKEN=secret-token
```

```toml
# tinker.toml
[database]
source = "env:DATABASE_URL"
type   = "postgres"

[api]
base_url  = "env:API_BASE_URL"
auth      = "env:API_TOKEN"
auth_type = "bearer"
```

### Auto-Detected Environment Variables

Tinker scans your `.env` files for database configuration. The following variables are recognized automatically:

| Variable | Inferred Type |
|----------|--------------|
| `DATABASE_URL`, `DB_URL` | Auto (by URL prefix) |
| `POSTGRES_URL` | PostgreSQL |
| `MYSQL_URL` | MySQL |
| `MONGO_URL`, `MONGODB_URI` | MongoDB |
| `DB_PATH`, `SQLITE_PATH`, `SQLITE_DB` | SQLite |
| `DB_HOST`, `DB_CONNECTION` | Auto (by value) |

File paths ending in `.db`, `.sqlite`, `.sqlite3` are automatically recognized as SQLite databases.

## Native Database Drivers

Tinker includes **pure Go database drivers** compiled into the binary — no external CLI tools needed for one-shot queries on PostgreSQL and MySQL:

| Driver | Package | Supports |
|--------|---------|----------|
| **PostgreSQL** | `jackc/pgx/v5` | All one-shot queries + `db explore` TUI |
| **MySQL** | `go-sql-driver/mysql` | All one-shot queries + `db explore` TUI |
| **SQLite** | `sqlite3`/`litecli`/`usql` | Queries via CLI tools |

**What this means in practice:**

- `tinker db tables`, `describe`, `count`, `find`, `exec`, `ping`, `size`, `explore` — all work for PostgreSQL and MySQL **without any external tools installed**
- SQLite uses CLI tools (`sqlite3`, `litecli`, or `usql`) for all queries — still fully supported, just not zero-dependency
- Cross-compilation works everywhere (pure Go, no CGO)
- External CLIs (`litecli`, `pgcli`, `mycli`, `usql`) are only needed for **interactive sessions** (`tinker db connect`)

## Database Explorer (TUI)

A full-screen interactive database browser built with [Bubble Tea](https://github.com/charmbracelet/bubbletea):

```bash
tinker db explore
```

| Key | Action |
|-----|--------|
| `↑`/`k`, `↓`/`j` | Navigate table list |
| `Enter` | View table data (up to 100 rows) |
| `s` | View `CREATE TABLE` statement |
| `Esc` / `Backspace` | Go back to table list |
| `q` | Quit |

Works with native drivers (PostgreSQL, MySQL) and CLI fallback (SQLite).

## Rich Output

Tinker renders professional-grade terminal output:

| Feature | Package | What it does |
|---------|---------|-------------|
| **Table rendering** | `jedib0t/go-pretty/v6` | Aligned tables for `describe`, `indexes`, `size`, `find`, `exec` |
| **Syntax highlighting** | `alecthomas/chroma/v2` | SQL highlighting for `db schema`, JSON highlighting for `api` responses |
| **JSON filtering** | `tidwall/gjson` | Native `--jq` filter — no `jq` binary needed for simple paths |
| **Styled UI** | `charmbracelet/lipgloss` | Colored badges, status indicators, visual hierarchy |

## Prerequisites

PostgreSQL and MySQL one-shot commands work out of the box with built-in drivers. For SQLite queries and **interactive sessions**, Tinker orchestrates companion tools. **`tinker deps install`** auto-installs the ones you need.

| Tool | Purpose | Required for | Install |
|------|---------|-------------|---------|
| **sqlite3** | SQLite queries | SQLite projects | `apt install sqlite3` / `brew install sqlite` |
| **litecli** | SQLite REPL (syntax highlighting) | Interactive | `pip install litecli` |
| **pgcli** | PostgreSQL REPL | Interactive | `pip install pgcli` |
| **mycli** | MySQL REPL | Interactive | `pip install mycli` |
| **usql** | Universal DB REPL | Interactive | `go install github.com/xo/usql@latest` |
| **curlie** | HTTP client | API feature | `go install github.com/rs/curlie@latest` |
| **evans** | gRPC REPL | gRPC feature | `go install github.com/ktr0731/evans@latest` |
| **grpcurl** | gRPC client | gRPC feature | `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` |

You don't need all of them — only the tools for features you use.

## Architecture

```
┌─────────────────────────────────────────────────┐
│              tinker CLI (pure Go)                │
├───────────────┬──────────────┬──────────────────┤
│  DB Module    │  API Module  │  gRPC Module     │
│  (pgx/mysql   │  (native +   │  (grpcurl/       │
│   + CLI)      │   CLI)       │   evans)         │
├───────────────┴──────────────┴──────────────────┤
│  OpenAPI Parser │ Commands │ Envs │ Docker      │
├─────────────────────────────────────────────────┤
│                Config Layer                      │
│         (tinker.toml + .env resolver)            │
├─────────────────────────────────────────────────┤
│           Auto-Detection Engine                  │
│          (tinker init scans project)             │
└─────────────────────────────────────────────────┘
```

**Design principles:**

- **Native first, CLI fallback** — Built-in pure Go drivers for PostgreSQL/MySQL queries; external CLIs for SQLite and interactive sessions
- **Compose, don't reimplement** — Shell out to best-in-class tools for interactive sessions instead of rewriting them
- **Contract is declarative** — `tinker.toml` is a simple config file, not a Go interface
- **Auto-detect first, configure second** — `tinker init` should work for 80% of cases
- **Cross-language by default** — Works with any project that has a `.env` and `tinker.toml`

## Comparison

| Feature | Laravel Tinker | gore | Tinker |
|---------|:-:|:-:|:-:|
| Styled TUI | ❌ | ❌ | ✅ lipgloss |
| Interactive DB session | ✅ Eloquent | ❌ | ✅ usql (raw SQL) |
| Call API endpoints | ✅ | ❌ | ✅ curlie |
| Call gRPC services | ❌ | ❌ | ✅ evans/grpcurl |
| Language-agnostic | ❌ PHP only | ❌ Go only | ✅ Any project |
| Framework-agnostic | ❌ Laravel only | ✅ | ✅ |
| Execute project code | ✅ Full PHP | ✅ Go | ⚠️ `tinker run` |
| Auto-configuration | ✅ | ❌ | ✅ `tinker init` |
| OpenAPI spec parsing | ❌ | ❌ | ✅ `tinker api endpoints` |
| Interactive API explorer | ❌ | ❌ | ✅ `tinker api explore` |
| Multi-environment | ❌ | ❌ | ✅ `tinker --env staging` |
| Custom commands | ❌ | ❌ | ✅ `tinker cmd <name>` |
| Built-in migrations | ❌ | ❌ | ✅ `tinker db migrate` |
| TUI database browser | ❌ | ❌ | ✅ `tinker db explore` |
| Docker Compose detection | ❌ | ❌ | ✅ `tinker docker` |

## Roadmap

- [x] Native database drivers (PostgreSQL, MySQL)
- [x] Built-in migrations with version tracking
- [x] Database seeding
- [x] Interactive TUI database browser (Bubble Tea)
- [x] OpenAPI spec parsing & interactive API explorer
- [x] Multi-environment support
- [x] Custom commands
- [x] Docker Compose detection
- [x] Shell completions (bash, zsh, fish, powershell)
- [x] Rich output (go-pretty tables, chroma syntax highlighting, gjson filtering)
- [ ] Plugin system for custom modules
- [ ] Native gRPC via grpcurl library (no external binary)
- [ ] HTTP session persistence (cookies, auth state across requests)

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes and add tests where applicable
4. Ensure linting passes (`golangci-lint run ./...`)
5. Ensure tests pass (`go test ./...`)
6. Commit with a descriptive message
7. Open a Pull Request

### Development

```bash
git clone https://github.com/mvaliolahi/tinker.git
cd tinker
go build ./cmd/tinker
go test ./...
golangci-lint run ./...
```

## License

[MIT](LICENSE) &copy; Mohammad Valiolahi
