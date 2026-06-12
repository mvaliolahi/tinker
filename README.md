# Tinker 🛠️

A project-aware CLI for database, API, and gRPC interaction — inspired by Laravel Tinker, built for Go (and any language).

Tinker comes with **built-in database drivers** (PostgreSQL, MySQL) for zero-dependency one-shot queries, and composes best-in-class open source tools (`usql`, `httpie`/`curlie`, `grpcurl`/`evans`) for interactive sessions. It reads your project's `tinker.toml` and `.env` to automatically configure connections, so you can jump straight into working with your project.

## Why Tinker?

If you've used Laravel Tinker, you know the workflow: open a shell, query the database, call an API endpoint, test a service — all within your project context. Go doesn't have an equivalent because it's compiled, not interpreted.

**Tinker takes a different approach.** Instead of trying to be a REPL for Go code, it:

1. **Reads your project's existing config** (`.env`, `tinker.toml`) — no wrapper abstractions
2. **Composes proven OSS tools** — `usql` for DB, `httpie` for HTTP, `evans`/`grpcurl` for gRPC
3. **Defines a simple contract** — any project that implements `tinker.toml` gets the full experience
4. **Works with any language** — not just Go! Any project with a `.env` and `tinker.toml` works

## Quick Start

### Install

**One-line install (recommended) — installs binary + configures PATH:**

```bash
curl -fsSL https://raw.githubusercontent.com/mvaliolahi/tinker/main/install.sh | bash
```

**Or manually with Go:**

```bash
go install github.com/mvaliolahi/tinker/cmd/tinker@latest
# Then add to PATH (if not already):
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Initialize

```bash
cd your-project
tinker init
```

Tinker will scan your project for:
- `.env` files with `DATABASE_URL`, `DB_PATH`, `API_BASE_URL`, `GRPC_ADDR`, etc.
- OpenAPI/Swagger spec files
- `.proto` directories
- SQLite database files (`.db`, `.sqlite`, `.sqlite3`)
- Docker Compose files

And generate a `tinker.toml`:

```toml
# tinker.toml
[database]
source = "env:DATABASE_URL"
type = "postgres"

[api]
base_url = "env:API_BASE_URL"
spec = "openapi.yaml"
auth = "env:API_TOKEN"
auth_type = "bearer"

[grpc]
addr = "env:GRPC_ADDR"
proto_dir = "./proto"
reflection = true
```

### Use

```bash
# Open interactive database session (usql)
tinker db

# Database queries (with handy shortcuts)
tinker db tables                              # list all tables (alias: ls)
tinker db describe users                      # show table schema (alias: desc)
tinker db indexes users                       # show table indexes (alias: idx)
tinker db schema users                        # show CREATE TABLE statement (alias: s)
tinker db count users                         # count rows (alias: c)
tinker db count users "status='active'"       # count with WHERE clause
tinker db find users 1                        # find row by ID (alias: f)
tinker db exec "SELECT * FROM users LIMIT 5"  # run SQL (aliases: e, sql)
tinker db ping                                # test database connectivity
tinker db size                                # show table row counts

# Database migrations
tinker db migrate up       # run pending migrations
tinker db migrate down     # rollback last migration
tinker db migrate status   # show migration status

# Seed the database
tinker db seed             # run all .sql files in seed/ directory
tinker db seed seed/users  # run a specific seed file or directory

# Interactive database browser (TUI)
tinker db explore          # open full-screen database browser

# Handy shortcuts
tinker db ls           # same as: tinker db tables
tinker db desc users   # same as: tinker db describe users
tinker db idx users    # same as: tinker db indexes users
tinker db s users      # same as: tinker db schema users
tinker db c users      # same as: tinker db count users
tinker db f users 1    # same as: tinker db find users 1
tinker db sql "SELECT" # same as: tinker db exec "SELECT 1"

# API endpoints (from OpenAPI spec)
tinker api endpoints                         # list all endpoints (alias: ep)
tinker api endpoints --tag users             # filter by tag
tinker api explore                           # interactive API explorer

# Call API endpoints
tinker api GET /users
tinker api POST /users '{"name": "Ali"}'
tinker api PUT /users/1 '{"name": "Updated"}'
tinker api DELETE /users/1

# gRPC interactions
tinker grpc                    # Interactive REPL (evans)
tinker grpc list               # List services
tinker grpc describe UserService
tinker grpc call UserService/GetUser '{"id": 1}'

# Custom commands (from [commands] section)
tinker cmd migrate             # run custom migrate command
tinker cmd seed                # run custom seed command
tinker cmd test                # run custom test command
tinker cmd list                # list available commands

# Multi-environment support
tinker --env staging db        # use staging environment
tinker --env production db ping
tinker env list                # list available environments
tinker env show staging        # show staging overrides

# Docker Compose
tinker docker                  # show Docker Compose info
tinker docker list             # list services (alias: ls)

# Configuration
tinker config show             # display resolved configuration
tinker config validate         # validate tinker.toml

# Run one-off Go code in project context
tinker run 'fmt.Println("Hello from tinker!")'

# Run Makefile targets
tinker make build
tinker make test
tinker make list               # list available targets

# Manage dependencies
tinker deps list               # check what's installed
tinker deps install            # install missing tools
```

## The Contract

The `tinker.toml` file is Tinker's contract with your project. It's declarative, minimal, and language-agnostic.

### Full Spec

```toml
[database]
# How to get the connection string:
#   "env:VAR_NAME" — read from environment variable (supports .env)
#   "postgres://user:pass@host:5432/db" — direct DSN
source = "env:DATABASE_URL"
# Database type: postgres, mysql, sqlite3, sqlserver, mongodb
type = "postgres"
# Override the driver name (optional)
driver = ""

[api]
# Base URL for API calls
base_url = "env:API_BASE_URL"
# Path to OpenAPI/Swagger spec (optional — enables autocomplete + explore)
spec = "openapi.yaml"
# Auth token source
auth = "env:API_TOKEN"
# Auth type: bearer, basic, api_key, or raw
auth_type = "bearer"
# Additional default headers
[api.headers]
X-Custom-Header = "value"

[grpc]
# gRPC server address
addr = "env:GRPC_ADDR"
# Directory containing .proto files
proto_dir = "./proto"
# Enable gRPC server reflection
reflection = true

# Custom project commands
[commands]
migrate = "go run ./cmd/migrate"
seed = "go run ./cmd/seed"
test = "go test ./..."

# Multi-environment overrides
[envs.staging.database]
source = "env:STAGING_DATABASE_URL"

[envs.staging.api]
base_url = "env:STAGING_API_BASE_URL"
auth = "env:STAGING_API_TOKEN"

[envs.production.database]
source = "env:PRODUCTION_DATABASE_URL"

[envs.production.api]
base_url = "env:PRODUCTION_API_BASE_URL"
auth = "env:PRODUCTION_API_TOKEN"
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
type = "postgres"

[api]
base_url = "env:API_BASE_URL"
auth = "env:API_TOKEN"
auth_type = "bearer"
```

## Native Database Drivers

Tinker v0.23+ includes **pure Go database drivers** compiled into the binary for PostgreSQL and MySQL:

| Driver | Package | Supports |
|--------|---------|----------|
| **PostgreSQL** | `jackc/pgx/v5` | All one-shot queries |
| **MySQL** | `go-sql-driver/mysql` | All one-shot queries |
| **SQLite** | `sqlite3`/`litecli`/`usql` | Queries via CLI tools |

**This means:**
- `tinker db ls`, `tinker db desc`, `tinker db c`, `tinker db f`, `tinker db sql` work for PostgreSQL and MySQL **without any external CLI tools installed**
- SQLite queries use CLI tools (`sqlite3`/`litecli`/`usql`) — all one-shot commands (`tables`, `describe`, `indexes`, `schema`, `count`, `find`, `exec`, `ping`, `size`) are fully supported
- Cross-compilation works everywhere (pure Go, no CGO)
- External CLIs (`litecli`, `pgcli`, `mycli`, `usql`) are needed for **interactive sessions** (`tinker db connect`)

### Auto-Detected Environment Variables

Tinker scans your `.env` files for database configuration. The following variables are detected automatically:

| Variable | Inferred Type |
|----------|--------------|
| `DATABASE_URL`, `DB_URL` | Auto (by URL prefix) |
| `POSTGRES_URL` | PostgreSQL |
| `MYSQL_URL` | MySQL |
| `MONGO_URL`, `MONGODB_URI` | MongoDB |
| `DB_PATH`, `SQLITE_PATH`, `SQLITE_DB` | SQLite |
| `DB_HOST`, `DB_CONNECTION` | Auto (by value) |

File paths ending in `.db`, `.sqlite`, `.sqlite3` are automatically recognized as SQLite databases.

## Database Migrations

Tinker v0.26+ includes a built-in migration system that tracks applied versions in a `_tinker_migrations` table.

### Migration File Format

Create SQL files in a `migrations/` directory with the naming convention `NNN_description.up.sql` / `NNN_description.down.sql`:

```
migrations/
  001_create_users.up.sql
  001_create_users.down.sql
  002_create_posts.up.sql
  002_create_posts.down.sql
```

### Commands

```bash
# Run all pending migrations
tinker db migrate up

# Rollback the last migration
tinker db migrate down

# Show migration status (applied vs pending)
tinker db migrate status
```

Tinker auto-detects your migrations directory by checking `migrations/`, `migrate/`, `db/migrations/`, `sql/migrations/`, `backend/migrations/`, and `backend/migrate/`.

## Database Seeding

Run SQL seed files to populate your database with initial or test data:

```bash
# Run all .sql files in seed/ directory (alphabetical order)
tinker db seed

# Run a specific seed file
tinker db seed seed/users.sql

# Run all .sql files in a directory
tinker db seed seed/
```

Seed files are split by semicolons (respecting quoted strings) and executed statement by statement.

## Database Explorer (TUI)

Tinker v0.26+ includes a full-screen interactive database browser built with [Bubble Tea](https://github.com/charmbracelet/bubbletea):

```bash
tinker db explore
```

Features:
- **Table list** — navigate with ↑/k ↓/j, press enter to view data
- **Data view** — scrollable table showing up to 100 rows with column headers
- **Schema view** — press `s` to see the CREATE TABLE statement
- **Navigation** — `esc` to go back, `q` to quit

## Rich Output

Tinker v0.24+ includes professional-grade output formatting:

| Feature | Package | What it does |
|---------|---------|-------------|
| **Table Rendering** | `jedib0t/go-pretty/v6` | Beautiful aligned tables for `db describe`, `db indexes`, `db size`, `db find`, `db exec` |
| **Syntax Highlighting** | `alecthomas/chroma/v2` | SQL highlighting for `db schema`, JSON highlighting for `api` responses, 500+ lexers |
| **JSON Filtering** | `tidwall/gjson` | Native `--jq` filter for `tinker api` — no `jq` binary needed for simple paths like `data.users.0.name` |

**Before (tab-separated):**
```
Column  Type    Nullable  Default  Key
id      INTEGER NOT NULL  NULL     PK
name    TEXT    NULL      NULL
```

**After (go-pretty table):**
```
COLUMN   TYPE     NULLABLE   DEFAULT   KEY
id       INTEGER  NOT NULL   NULL      PK
name     TEXT     NULL       NULL
```

The `--jq` flag now tries **gjson** first (native Go), then falls back to `jq` CLI for complex expressions:
```bash
# Simple path — gjson (no jq binary needed)
tinker api GET /users -q "data.0.name"

# Complex expression — falls back to jq CLI
tinker api GET /users -q '.[] | select(.active)'
```

## OpenAPI Spec Integration

Tinker v0.25+ can parse your OpenAPI/Swagger spec to provide endpoint discovery and an interactive API explorer.

### List Endpoints

```bash
# List all endpoints from your spec
tinker api endpoints

# Filter by tag
tinker api endpoints --tag users

# Shortcut
tinker api ep
```

Output:
```
  API  Endpoints
  spec: My API v2.1.0
  total: 12 endpoint(s)

  users
    GET    /users                         — List all users
    POST   /users                         — Create a user
    GET    /users/{id}                    — Get user by ID
    PUT    /users/{id}                    — Update user
    DELETE /users/{id}                    — Delete user
```

### Interactive API Explorer

```bash
tinker api explore
```

Provides a REPL for browsing and calling endpoints:
```
  API  API Explorer
  base: http://localhost:8080
  spec: My API v2.1.0
  endpoints: 12 available

  Commands:
    list              List all endpoints
    call <method> <path> [body]   Call an endpoint
    find <keyword>    Search endpoints
    tags              List tags
    quit / q          Exit

tinker/api> GET /users/1
```

### Spec Configuration

Add the `spec` field to your `[api]` section:

```toml
[api]
base_url = "env:API_BASE_URL"
spec = "openapi.yaml"    # or swagger.json, openapi.yml
auth = "env:API_TOKEN"
auth_type = "bearer"
```

Supports both YAML and JSON formats (OpenAPI 3.x and Swagger 2.x).

## Multi-Environment Support

Tinker v0.25+ supports environment-specific configuration overrides via the `[envs.*]` sections in `tinker.toml`.

### Configuration

```toml
# Base configuration (default environment)
[database]
source = "env:DATABASE_URL"
type = "postgres"

[api]
base_url = "env:API_BASE_URL"
auth = "env:API_TOKEN"
auth_type = "bearer"

# Staging overrides
[envs.staging.database]
source = "env:STAGING_DATABASE_URL"

[envs.staging.api]
base_url = "env:STAGING_API_BASE_URL"
auth = "env:STAGING_API_TOKEN"

# Production overrides
[envs.production.database]
source = "env:PRODUCTION_DATABASE_URL"

[envs.production.api]
base_url = "env:PRODUCTION_API_BASE_URL"
auth = "env:PRODUCTION_API_TOKEN"
```

### Usage

```bash
# Default environment
tinker db ping

# Staging environment
tinker --env staging db ping

# Production environment
tinker --env production db tables

# List available environments
tinker env list

# Show overrides for an environment
tinker env show staging
```

Only the fields you specify in the override section are changed; everything else inherits from the base configuration.

## Custom Commands

Tinker v0.25+ supports custom project commands via the `[commands]` section in `tinker.toml`.

### Configuration

```toml
[commands]
migrate = "go run ./cmd/migrate"
seed = "go run ./cmd/seed"
test = "go test ./..."
lint = "golangci-lint run"
dev = "air"
build = "go build -o bin/app ./cmd/app"
```

### Usage

```bash
# Run a custom command
tinker cmd migrate
tinker cmd seed
tinker cmd test

# With extra arguments (appended to the command)
tinker cmd test -v -run TestUser

# List available commands
tinker cmd list
```

Commands are executed via your system shell, so they can include pipes, redirects, and any shell features.

## Docker Compose Integration

Tinker v0.25+ auto-detects Docker Compose files and extracts service information.

### Usage

```bash
# Show Docker Compose info
tinker docker

# List services
tinker docker list
```

Output:
```
  Docker Compose Services
  file: docker-compose.yml

  db                   postgres:15       ports: 5432:5432
    detected: [DB database]
  api                  myapp:latest      ports: 8080:8080
    detected: [API api]
```

### Auto-Detection

`tinker init` automatically detects:
- `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`
- Environment-specific compose files (`docker-compose.staging.yml`, etc.)
- Service types (database, API, gRPC) based on image names, service names, and environment variables

## Prerequisites

Tinker's one-shot DB commands for PostgreSQL and MySQL work out of the box with built-in drivers. SQLite requires a CLI tool (`sqlite3`/`litecli`/`usql`) for all queries. For **interactive sessions**, Tinker orchestrates existing tools. **`tinker init` auto-installs the ones you need** based on what's detected in your project. You can also manage them manually:

```bash
# Check which tools are installed
tinker deps list

# Install all missing tools
tinker deps install
```

| Tool | Purpose | Required? | Manual Install |
|------|---------|-----------|----------|
| **sqlite3** | SQLite queries + REPL | SQLite projects | System package (`apt install sqlite3`, `brew install sqlite`) |
| **litecli** | SQLite REPL (syntax highlighting) | Interactive only | `pip install litecli` |
| **pgcli** | PostgreSQL REPL (syntax highlighting) | Interactive only | `pip install pgcli` |
| **mycli** | MySQL REPL (syntax highlighting) | Interactive only | `pip install mycli` |
| **usql** | Database REPL (universal fallback) | Interactive only | `go install github.com/xo/usql@latest` |
| **httpie** | HTTP client | API feature | `pip install httpie` or `brew install httpie` |
| **curlie** | HTTP client (alt) | API feature | `go install github.com/rs/curlie@latest` |
| **evans** | gRPC REPL | gRPC feature | `go install github.com/ktr0731/evans@latest` |
| **grpcurl** | gRPC client | gRPC feature | `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` |

You don't need all of them — only install the tools for the features you use.

## Architecture

```
┌─────────────────────────────────────────┐
│            tinker CLI (Go)              │
├──────────────┬───────────┬──────────────┤
│  DB Module   │ API Module│ gRPC Module  │
│  (native +   │ (native + │ (grpcurl/    │
│   CLI)       │  CLI)     │  evans)      │
├──────────────┴───────────┴──────────────┤
│  OpenAPI Parser │ Commands │ Envs │ Docker│
├─────────────────────────────────────────┤
│           Config Layer                  │
│    (tinker.toml + .env resolver)        │
├─────────────────────────────────────────┤
│        Auto-Detection Engine            │
│    (tinker init scans project)          │
└─────────────────────────────────────────┘
```

**Design principles:**
- **Native first, CLI fallback** — Built-in pure Go drivers for PostgreSQL/MySQL queries; external CLIs for SQLite and interactive sessions
- **Compose, don't reimplement** — Shell out to httpie/evans for interactive sessions, don't rewrite them
- **Contract is declarative** — `tinker.toml`, not a Go interface
- **Auto-detect first, configure second** — `tinker init` should work for 80% of cases
- **Cross-language by default** — Works with any project that has a `.env` and `tinker.toml`
- **Beautiful terminal UI** — Styled output with [lipgloss](https://github.com/charmbracelet/lipgloss), colored badges, and clear visual hierarchy
- **Spec-driven API exploration** — Parse OpenAPI specs for endpoint discovery and interactive exploration

## Comparison

| Feature | Laravel Tinker | gore | Tinker (this project) |
|---------|---------------|------|----------------------|
| Styled TUI | ❌ | ❌ | ✅ lipgloss |
| Interactive DB session | ✅ Eloquent | ❌ | ✅ usql (raw SQL) |
| Call API endpoints | ✅ HTTP client | ❌ | ✅ httpie/curlie |
| Call gRPC services | ❌ | ❌ | ✅ evans/grpcurl |
| Language-agnostic | ❌ PHP only | ❌ Go only | ✅ Any project |
| Framework-agnostic | ❌ Laravel only | ✅ | ✅ |
| Execute project code | ✅ Full PHP | ✅ Go | ⚠️ `tinker run` (limited) |
| Auto-configuration | ✅ | ❌ | ✅ `tinker init` |
| Persistent state | ✅ | ⚠️ Per eval | ✅ DB session |
| OpenAPI spec parsing | ❌ | ❌ | ✅ `tinker api endpoints` |
| Interactive API explorer | ❌ | ❌ | ✅ `tinker api explore` |
| Multi-environment | ❌ | ❌ | ✅ `tinker --env staging` |
| Custom commands | ❌ | ❌ | ✅ `tinker cmd <name>` |
| Docker Compose detection | ❌ | ❌ | ✅ `tinker docker` |

## Roadmap

- [x] Native database drivers (PostgreSQL, MySQL) — no external CLI needed for queries
- [x] `tinker db index` — show table indexes
- [x] `tinker db ping` — test connectivity
- [x] `tinker db size` — table row counts
- [x] Rich table formatting for describe/indexes output
- [x] Shell completion (bash, zsh, fish, powershell)
- [x] Rich table rendering (go-pretty)
- [x] Syntax highlighting for SQL output (chroma)
- [x] Native JSON filtering via gjson (--jq without jq binary)
- [x] OpenAPI spec parsing for API endpoint autocomplete
- [x] `tinker api explore` — interactive API explorer
- [x] Custom command support from `[commands]` section
- [x] Multi-environment support (`tinker --env staging db`)
- [x] Docker integration — auto-detect docker-compose services
- [x] `tinker db seed` — run seed files (SQL)
- [x] `tinker db migrate` — run migrations (up/down/status with version tracking)
- [x] `tinker db explore` — interactive TUI database browser (Bubble Tea)
- [ ] Plugin system for custom modules
- [ ] Native gRPC via grpcurl library (no external binary)
- [ ] HTTP session persistence (cookies, auth state across requests)

## Contributing

Contributions are welcome! This project is in early development.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License — see [LICENSE](LICENSE) for details.
