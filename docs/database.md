# Database

Tinker provides a comprehensive set of database commands for introspection, querying, migration, and interactive exploration. For PostgreSQL and MySQL, all one-shot commands work natively via built-in pure Go drivers — no external CLI tools required.

## Native Database Drivers

Tinker includes pure Go database drivers compiled into the binary, which means one-shot queries work without any external tools installed:

| Driver | Package | Supports |
|--------|---------|----------|
| **PostgreSQL** | `jackc/pgx/v5` | All one-shot queries + `db explore` TUI |
| **MySQL** | `go-sql-driver/mysql` | All one-shot queries + `db explore` TUI |
| **SQLite** | CLI fallback | Queries via `sqlite3`, `litecli`, or `usql` |

**What this means in practice:**

- Commands like `db tables`, `describe`, `count`, `find`, `exec`, `ping`, `size`, and `explore` all work for PostgreSQL and MySQL **without any external tools installed**
- SQLite uses CLI tools for all queries — still fully supported, just not zero-dependency
- External CLIs (`litecli`, `pgcli`, `mycli`, `usql`) are only needed for **interactive sessions** (`tinker db connect`)
- The binary is pure Go with no CGO dependency, so cross-compilation works everywhere

## Commands

### List Tables

```bash
tinker db tables
# Alias: tinker db ls
```

Lists all tables in the connected database. For PostgreSQL and MySQL, this uses a native query against `information_schema`. For SQLite, it falls back to the CLI.

### Describe Table

```bash
tinker db describe <table>
# Alias: tinker db desc <table>
```

Shows the column schema for a specific table, including column name, data type, nullable, and default value. The output is rendered as a formatted table using go-pretty.

Example output:

```
  DB  Describe users

  Column     Type         Nullable  Default
  ─────────  ──────────   ────────  ───────
  id         integer      NO        nextval(...)
  name       varchar(255) NO
  email      varchar(255) NO
  created_at  timestamp    YES       now()
```

### Show Indexes

```bash
tinker db indexes <table>
# Alias: tinker db idx <table>
```

Shows all indexes on a table, including the index name, columns, uniqueness, and type. For SQLite, this joins `pragma_index_list` and `pragma_index_info` to provide column names. For MySQL, it queries `information_schema.statistics`.

### Show Schema

```bash
tinker db schema <table>
# Alias: tinker db s <table>
```

Displays the `CREATE TABLE` statement for a table. The output is syntax-highlighted using Chroma with the Monokai theme, making it easy to read complex schemas.

### Count Rows

```bash
tinker db count <table> [where]
# Alias: tinker db c <table> [where]
```

Counts rows in a table. You can optionally provide a `WHERE` clause:

```bash
tinker db count users                    # total count
tinker db count users "active = true"    # filtered count
```

### Find Row by ID

```bash
tinker db find <table> <id>
# Alias: tinker db f <table> <id>
```

Finds a single row by its ID column. The result is rendered as a formatted table with column names as headers.

```bash
tinker db find users 1
```

### Execute SQL

```bash
tinker db exec "<sql>"
# Aliases: tinker db e "<sql>", tinker db sql "<sql>"
```

Runs arbitrary SQL statements. The command tries the native connection first (for PostgreSQL and MySQL), then falls back to the CLI if the native query fails. Results are rendered as formatted tables.

```bash
tinker db exec "SELECT * FROM users WHERE active = true LIMIT 10"
tinker db exec "INSERT INTO users (name, email) VALUES ('Ali', 'ali@example.com')"
tinker db exec "UPDATE users SET active = false WHERE id = 5"
```

> **Note:** For data-modifying statements (INSERT, UPDATE, DELETE), the output depends on the database driver. Some drivers return affected row counts; others return empty results.

### Ping Database

```bash
tinker db ping
```

Tests database connectivity and measures response time. Useful for verifying your connection configuration and diagnosing network issues.

```
  DB  ✓ reachable  3ms
```

### Show Table Sizes

```bash
tinker db size
```

Shows row counts for all tables in the database. This is useful for getting a quick overview of data distribution across your tables.

### Connect (Interactive Session)

```bash
tinker db connect
```

Opens an interactive database session using the best available CLI tool:

| Database | Preferred CLI | Fallback |
|----------|--------------|----------|
| PostgreSQL | `pgcli` (syntax highlighting + autocomplete) | `usql` |
| MySQL | `mycli` (syntax highlighting + autocomplete) | `usql` |
| SQLite | `litecli` (syntax highlighting + autocomplete) | `sqlite3` → `usql` |

If the preferred CLI is not installed, Tinker suggests how to install it and falls back to the next available tool.

### Explore (TUI Browser)

```bash
tinker db explore
```

Opens a full-screen interactive database browser built with [Bubble Tea](https://github.com/charmbracelet/bubbletea). This is the most powerful way to browse your database interactively.

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `↑`/`k`, `↓`/`j` | Navigate table list |
| `Enter` | View table data (up to 100 rows) |
| `s` | View `CREATE TABLE` statement |
| `Esc` / `Backspace` | Go back to table list |
| `q` | Quit |

The explore TUI works with native drivers (PostgreSQL, MySQL) and CLI fallback (SQLite). It uses the same native connections as other commands, so no additional setup is needed.

## Migrations

Tinker includes a built-in migration system that tracks applied versions in a `_tinker_migrations` table. This table is automatically created on first use and stores the version number, migration name, and application timestamp.

### Migration file format

Migration files follow the `NNN_description.up.sql` / `NNN_description.down.sql` convention:

```
migrations/
  001_create_users.up.sql
  001_create_users.down.sql
  002_create_posts.up.sql
  002_create_posts.down.sql
  003_add_email_index.up.sql
  003_add_email_index.down.sql
```

- The version number (`001`, `002`, etc.) determines the execution order
- The description is for human readability
- `.up.sql` files are run during `migrate up`
- `.down.sql` files are run during `migrate down` (rollback)
- SQL files are split by semicolons, respecting quoted strings and `--` line comments

### Run migrations

```bash
tinker db migrate up      # run all pending migrations
```

This command:
1. Finds the migration directory (auto-detected or configured via `migrate_dir`)
2. Creates the `_tinker_migrations` table if it doesn't exist
3. Checks which migrations have already been applied
4. Runs all pending migrations in version order
5. Records each applied migration in the tracking table

### Rollback migrations

```bash
tinker db migrate down
# Alias: tinker db migrate rollback
```

Rolls back the most recently applied migration by running its `.down.sql` file and removing the record from the tracking table. Only one migration is rolled back per invocation (to roll back multiple, run the command multiple times).

### Check migration status

```bash
tinker db migrate status
# Alias: tinker db migrate st
```

Shows the status of all migrations:

```
  DB  Migration Status

  Version  Name             Status
  ───────  ──────────────   ──────────
  001      create_users     ✓ applied
  002      create_posts     ✓ applied
  003      add_email_index  ⚠ pending
```

### Migration directory detection

Tinker auto-detects migration directories by checking for the following paths (in order):

1. Configured `migrate_dir` in `tinker.toml` (if set)
2. `migrations/`
3. `db/migrations/`
4. `backend/migrations/`
5. `sql/migrations/`

You can override this by setting `migrate_dir` in your `tinker.toml`:

```toml
[database]
migrate_dir = "db/migrations"
```

## Seeding

Tinker can execute SQL seed files against your database. Seed files are split by semicolons (respecting quoted strings and `--` comments) and executed statement by statement.

### Run seed files

```bash
tinker db seed                # run all .sql files in seed/ directory
tinker db seed seed/users.sql # run a specific seed file
tinker db seed fixtures/      # run all .sql files in a directory
```

### Seed directory detection

Similar to migrations, seed directories are auto-detected:

1. Configured `seed_dir` in `tinker.toml` (if set)
2. `seed/`
3. `db/seed/`
4. `db/seeds/`
5. `fixtures/`

You can override this by setting `seed_dir` in your `tinker.toml`:

```toml
[database]
seed_dir = "db/seed"
```

## Rich Output

All database query output is rendered with professional-grade terminal formatting:

| Feature | What it does |
|---------|-------------|
| **go-pretty tables** | Aligned, bordered tables for `describe`, `indexes`, `size`, `find`, `exec` |
| **Chroma highlighting** | SQL syntax highlighting for `db schema` output (Monokai theme, 256-color) |
| **Lipgloss styling** | Colored badges, status indicators, visual hierarchy |

The output automatically adapts to your terminal's color support. If your terminal doesn't support 256 colors, the output gracefully degrades to simpler formatting.

## Connection Architecture

Tinker uses a two-tier connection strategy:

1. **Native connection** (`database/sql`) — Used for PostgreSQL and MySQL. Opened when a `Session` is created, verified with a `Ping()`, and closed when the session ends. This handles all one-shot queries efficiently without spawning external processes.

2. **CLI fallback** — Used for SQLite and when native connections fail. Spawns external tools (`sqlite3`, `usql`, etc.) as subprocesses and captures their output. Required for interactive sessions (`db connect`) because native Go drivers don't provide REPL functionality.

The `HasNativeConn()` method on `Session` reports whether a native connection is available. Commands like `db exec` try native first and fall back to CLI automatically, so you always get the best available experience.
