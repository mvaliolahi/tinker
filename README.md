# Tinker 🛠️

A project-aware CLI for database, API, and gRPC interaction — inspired by Laravel Tinker, built for Go (and any language).

Tinker composes best-in-class open source tools (`usql`, `httpie`/`curlie`, `grpcurl`/`evans`) into a unified interface. It reads your project's `tinker.toml` and `.env` to automatically configure connections, so you can jump straight into working with your project.

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
- `.env` files with `DATABASE_URL`, `API_BASE_URL`, `GRPC_ADDR`, etc.
- OpenAPI/Swagger spec files
- `.proto` directories
- SQLite database files

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

# Quick database queries
tinker db tables
tinker db describe users
tinker db count users
tinker db find users 1
tinker db exec "SELECT * FROM users LIMIT 5"

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

# Run one-off Go code in project context
tinker run 'fmt.Println("Hello from tinker!")'
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
# Path to OpenAPI/Swagger spec (optional — enables autocomplete)
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

## Prerequisites

Tinker orchestrates existing tools. **`tinker init` auto-installs the ones you need** based on what's detected in your project. You can also manage them manually:

```bash
# Check which tools are installed
tinker deps list

# Install all missing tools
tinker deps install
```

| Tool | Purpose | Manual Install |
|------|---------|---------|
| **usql** | Database REPL | `go install github.com/xo/usql@latest` |
| **httpie** | HTTP client | `pip install httpie` or `brew install httpie` |
| **curlie** | HTTP client (alt) | `go install github.com/rs/curlie@latest` |
| **evans** | gRPC REPL | `go install github.com/ktr0731/evans@latest` |
| **grpcurl** | gRPC client | `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` |

You don't need all of them — only install the tools for the features you use.

> **Note:** If `go install` fails due to proxy rate limits, try `GOPROXY=direct go install <module>@latest`.

## Architecture

```
┌─────────────────────────────────────────┐
│            tinker CLI (Go)              │
├──────────────┬───────────┬──────────────┤
│  DB Module   │ API Module│ gRPC Module  │
│  (usql)      │ (httpie)  │ (evans/      │
│              │ /curlie)  │  grpcurl)    │
├──────────────┴───────────┴──────────────┤
│           Config Layer                  │
│    (tinker.toml + .env resolver)        │
├─────────────────────────────────────────┤
│        Auto-Detection Engine            │
│    (tinker init scans project)          │
└─────────────────────────────────────────┘
```

**Design principles:**
- **Compose, don't reimplement** — Shell out to usql/httpie/evans, don't rewrite them
- **Contract is declarative** — `tinker.toml`, not a Go interface
- **Auto-detect first, configure second** — `tinker init` should work for 80% of cases
- **Cross-language by default** — Works with any project that has a `.env` and `tinker.toml`

## Comparison

| Feature | Laravel Tinker | gore | Tinker (this project) |
|---------|---------------|------|----------------------|
| Interactive DB session | ✅ Eloquent | ❌ | ✅ usql (raw SQL) |
| Call API endpoints | ✅ HTTP client | ❌ | ✅ httpie/curlie |
| Call gRPC services | ❌ | ❌ | ✅ evans/grpcurl |
| Language-agnostic | ❌ PHP only | ❌ Go only | ✅ Any project |
| Framework-agnostic | ❌ Laravel only | ✅ | ✅ |
| Execute project code | ✅ Full PHP | ✅ Go | ⚠️ `tinker run` (limited) |
| Auto-configuration | ✅ | ❌ | ✅ `tinker init` |
| Persistent state | ✅ | ⚠️ Per eval | ✅ DB session |

## Roadmap

- [ ] OpenAPI spec parsing for API endpoint autocomplete
- [ ] `tinker api explore` — interactive API explorer
- [ ] Custom command support from `[commands]` section
- [ ] Connection testing (`tinker db ping`, `tinker api ping`)
- [ ] Multi-environment support (`tinker --env staging db`)
- [ ] Plugin system for custom modules
- [ ] Shell completion (bash, zsh, fish)
- [ ] `tinker db seed` — run seed files
- [ ] `tinker db migrate` — run migrations
- [ ] Docker integration — auto-detect docker-compose services

## Contributing

Contributions are welcome! This project is in early development.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License — see [LICENSE](LICENSE) for details.
