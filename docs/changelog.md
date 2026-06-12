# Changelog

All notable changes to Tinker are documented in this file. The project follows [semantic versioning](https://semver.org/).

## v0.28.0

- Bumped version for new update tag

## v0.27.0

- Implemented plugin system with script plugins and lifecycle hooks
- Added native gRPC support via grpcurl library (no external binary for list/describe)
- Added HTTP session persistence (cookies, auth state across requests)

## v0.26.0 — v0.26.2

- Added database migration system (`db migrate up`, `db migrate down`, `db migrate status`)
- Added database seeding (`db seed`)
- Added interactive TUI database browser (`db explore`) built with Bubble Tea
- Bug fixes and stability improvements

## v0.25.0 — v0.25.5

- Added OpenAPI spec parsing and interactive API explorer
- Added multi-environment support (`tinker --env staging`)
- Added custom commands (`tinker cmd <name>`)
- Added Docker Compose detection and inspection
- Added environment commands (`tinker env list`, `tinker env show`)
- Added custom command execution via system shell
- Bug fixes for db size, SQLite compatibility, and golangci-lint errors

## v0.24.0

- Added rich output: go-pretty tables for all database query results
- Added Chroma syntax highlighting for SQL (db schema) and JSON (api responses)
- Added gjson-based `--jq` flag for native JSON response filtering (no jq binary needed for simple paths)
- External jq binary still used as fallback for complex expressions

## v0.23.0

- Added native database drivers: PostgreSQL (`jackc/pgx/v5`), MySQL (`go-sql-driver/mysql`)
- All one-shot database commands now work without external CLI tools for PostgreSQL and MySQL
- External CLIs only required for interactive sessions (`db connect`) and SQLite queries
- Eliminated CLI execution bugs (sqlite3 -c flag, DSN format mismatches, file: prefix issues)

## v0.22.0

- Initial public release with core database, API, and gRPC commands
- Auto-detection engine for project configuration
- `tinker init` command for zero-config setup
- Dashboard with project overview
