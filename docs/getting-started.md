# Getting Started

Welcome to Tinker — a project-aware CLI for database, API, and gRPC interaction. This guide walks you through installing Tinker, initializing it in your project, and running your first commands.

## What is Tinker?

Tinker is a command-line tool inspired by Laravel Tinker, designed for developers who need to quickly interact with their project's database, HTTP APIs, and gRPC services. Unlike Laravel Tinker (which is PHP-only), Tinker is **language-agnostic** — it works with any project that has a `.env` file and a `tinker.toml` configuration.

Key capabilities include:

- **Built-in database drivers** for PostgreSQL and MySQL (zero external dependencies for one-shot queries)
- **OpenAPI spec parsing** with an interactive API explorer
- **Native gRPC** service listing and description via server reflection
- **Built-in migrations** with version tracking
- **Interactive TUI** database browser built with Bubble Tea
- **Multi-environment support** (staging, production, etc.)
- **Plugin system** for custom scripts and lifecycle hooks
- **Docker Compose detection** and service inspection

## Installation

### One-line install (recommended)

The easiest way to install Tinker is via the install script, which downloads the binary and configures your PATH:

```bash
curl -fsSL https://raw.githubusercontent.com/mvaliolahi/tinker/main/install.sh | bash
```

This script detects your operating system and architecture, downloads the appropriate binary, and places it in `$HOME/.local/bin` (or `$HOME/go/bin` as a fallback). After installation, restart your shell or run `source ~/.bashrc` (or the equivalent for your shell) to update your PATH.

### Install with Go

If you have Go 1.25 or later installed, you can build Tinker directly from source:

```bash
go install github.com/mvaliolahi/tinker/cmd/tinker@latest
```

The binary will be placed in `$GOPATH/bin` (typically `$HOME/go/bin`). Make sure this directory is in your `PATH`:

```bash
export PATH="$PATH:$HOME/go/bin"
```

### Homebrew

For macOS users, Tinker is available via Homebrew:

```bash
brew tap mvaliolahi/tap
brew install tinker
```

### Verify installation

After installing, verify that Tinker is available:

```bash
tinker version
```

You should see output like:

```
╔═══════════════════════╗
║  ⚡ Tinker v0.28.0  ║
╚═══════════════════════╝
```

## Prerequisites

Tinker is designed to work with minimal dependencies. For PostgreSQL and MySQL projects, **no external tools are required** for one-shot queries — the database drivers are compiled into the binary.

For SQLite projects and **interactive sessions**, Tinker orchestrates companion tools. You can install them automatically with `tinker deps install`.

| Tool | Purpose | Required for | Install |
|------|---------|-------------|---------|
| **sqlite3** | SQLite queries | SQLite projects | `apt install sqlite3` / `brew install sqlite` |
| **litecli** | SQLite REPL (syntax highlighting) | Interactive | `pip install litecli` |
| **pgcli** | PostgreSQL REPL | Interactive | `pip install pgcli` |
| **mycli** | MySQL REPL | Interactive | `pip install mycli` |
| **usql** | Universal DB REPL | Interactive | `go install github.com/xo/usql@latest` |
| **curlie** | HTTP client | API feature | `go install github.com/rs/curlie@latest` |
| **evans** | gRPC REPL | gRPC feature | `go install github.com/ktr0731/evans@latest` |
| **grpcurl** | gRPC client | gRPC call command | `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` |

You do not need all of them — only the tools for features you actually use. The `tinker deps` command helps you check and install what you need:

```bash
tinker deps list       # check which tools are installed
tinker deps install    # auto-install missing tools
```

## Quick Start

### Step 1: Navigate to your project

```bash
cd your-project
```

Tinker works in the context of your project directory. It looks for a `tinker.toml` configuration file and `.env` files in the current directory (or any parent directory up to the filesystem root).

### Step 2: Initialize

```bash
tinker init
```

Tinker scans your project and automatically detects:

- `.env` files with `DATABASE_URL`, `DB_PATH`, `API_BASE_URL`, `GRPC_ADDR`, etc.
- OpenAPI / Swagger spec files (`openapi.yaml`, `swagger.json`, and more)
- `.proto` directories for gRPC services
- SQLite database files (`.db`, `.sqlite`, `.sqlite3`)
- Docker Compose files (`docker-compose.yml`, `compose.yaml`, etc.)
- Migration and seed directories

Based on what it finds, Tinker generates a `tinker.toml` configuration file and installs any required companion tools. The init process has four steps:

1. **Scanning** — looks for `.env`, docker-compose, and spec files
2. **Detecting** — identifies database, API, gRPC, log, and Docker services
3. **Installing** — auto-installs companion tools for detected services
4. **Ready** — prints next-step hints

### Step 3: Run commands

Once initialized, you can immediately start working with your project:

```bash
# Database
tinker db tables           # list all tables (native — no CLI tools needed)
tinker db describe users   # show column schema for the users table
tinker db exec "SELECT * FROM users LIMIT 5"  # run arbitrary SQL
tinker db explore          # full-screen TUI database browser

# API
tinker api endpoints       # list endpoints from your OpenAPI spec
tinker api GET /users      # call an API endpoint
tinker api explore         # interactive API explorer

# gRPC
tinker grpc list           # list gRPC services (native — no grpcurl needed)
tinker grpc describe UserService  # describe a service

# Dashboard
tinker                     # show project overview with detected services
```

## Next Steps

- [Configuration](configuration.md) — learn about `tinker.toml` options, environment variables, and multi-environment setup
- [Database](database.md) — deep dive into database commands, migrations, seeding, and the TUI browser
- [API](api.md) — HTTP client usage, OpenAPI spec parsing, session persistence, and the interactive explorer
- [gRPC](grpc.md) — service listing, description, and method invocation
- [CLI Reference](cli-reference.md) — complete list of commands, flags, and aliases
- [Architecture](architecture.md) — internal design, project structure, and how the pieces fit together
