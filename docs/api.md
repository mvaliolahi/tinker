# API

Tinker provides a powerful HTTP client for interacting with your project's API endpoints. It supports direct endpoint calls, OpenAPI spec parsing with an interactive explorer, session persistence, and response filtering.

## Quick Reference

```bash
tinker api GET /users                          # call an endpoint
tinker api POST /users '{"name": "Ali"}'       # POST with JSON body
tinker api PUT /users/1 '{"name": "Updated"}'  # PUT with JSON body
tinker api DELETE /users/1                     # DELETE request

tinker api GET /users -q "data.0.name"         # gjson path filter
tinker api GET /users -q '.[] | select(.active)'  # complex jq filter

tinker api endpoints                           # list endpoints from spec
tinker api endpoints --tag users               # filter by tag
tinker api explore                             # interactive API explorer

tinker api session show                        # view session state
tinker api session clear                       # clear session state
```

## Calling Endpoints

### Basic usage

```bash
tinker api <method> <path> [body]
```

- If no method is provided, defaults to `GET`
- The `path` is appended to the configured `base_url`
- The `body` is sent as-is (typically JSON)

### Examples

```bash
# Simple GET
tinker api GET /users

# GET with a single argument (method defaults to GET)
tinker api /users

# POST with JSON body
tinker api POST /users '{"name": "Ali", "email": "ali@example.com"}'

# PUT to update a resource
tinker api PUT /users/1 '{"name": "Updated Name"}'

# DELETE a resource
tinker api DELETE /users/1
```

### Authentication

Tinker supports four authentication types, configured in `tinker.toml`:

| Auth Type | Config Value | Header Sent |
|-----------|-------------|-------------|
| Bearer Token | `auth_type = "bearer"` | `Authorization: Bearer <token>` |
| Basic Auth | `auth_type = "basic"` | `Authorization: Basic <token>` |
| API Key | `auth_type = "api_key"` | `X-API-Key: <token>` |
| Raw | `auth_type = "raw"` | `Authorization: <token>` |

The `auth` value supports `env:` prefix for environment variable resolution:

```toml
[api]
base_url  = "env:API_BASE_URL"
auth      = "env:API_TOKEN"
auth_type = "bearer"
```

### Custom Headers

You can define default headers that are sent with every request:

```toml
[api.headers]
X-Custom-Header = "value"
X-Request-ID = "auto-generated"
```

Headers from the `[api.headers]` section are merged with authentication headers on every request.

## Response Filtering

Tinker supports two levels of JSON response filtering via the `--jq` / `-q` flag:

### gjson path filtering (native, no external tools)

For simple dot-notation paths, Tinker uses [gjson](https://github.com/tidwall/gjson) natively — no `jq` binary required:

```bash
tinker api GET /users -q "data.0.name"           # first user's name
tinker api GET /users -q "data.#.name"            # all user names (array)
tinker api GET /users -q "data.#(active==true)"   # filter active users
```

Common gjson path patterns:

| Pattern | Description |
|---------|-------------|
| `data.0.name` | Access first element's name |
| `data.#` | Count of array elements |
| `data.#.name` | Extract all `name` fields from array |
| `data.#(active==true)` | Filter objects where `active` is true |
| `users.0.posts.0.title` | Nested path access |

### jq expression filtering (requires jq CLI)

For complex transformations, Tinker falls back to the `jq` CLI:

```bash
tinker api GET /users -q '.[] | select(.active) | .name'    # complex jq
tinker api GET /users -q '[.[] | {name, email}]'            # reshape JSON
tinker api GET /users -q 'group_by(.role) | map({role: .[0].role, count: length})'
```

Tinker tries gjson first. If the expression doesn't match gjson syntax, it falls back to `jq`. This means simple paths work without `jq` installed, while complex expressions require the `jq` binary.

## OpenAPI Spec Parsing

Tinker can parse OpenAPI/Swagger specification files (both YAML and JSON formats) and use them to discover and explore your API endpoints.

### List endpoints

```bash
tinker api endpoints
# Alias: tinker api ep
```

Lists all endpoints found in your OpenAPI spec, grouped by tag:

```
  API  Endpoints
  spec: My API v1.0.0
  total: 12 endpoint(s)

  ▪ users
    GET    /users          List users
    POST   /users          Create user
    GET    /users/{id}     Get user
    PUT    /users/{id}     Update user
    DELETE /users/{id}     Delete user

  ▪ auth
    POST   /auth/login     Login
    POST   /auth/logout    Logout
```

### Filter by tag

```bash
tinker api endpoints --tag users
# Short: tinker api endpoints -t users
```

Shows only endpoints belonging to the specified OpenAPI tag. If the tag doesn't exist, Tinker lists the available tags.

### Configuration

To use OpenAPI spec features, add the `spec` option to your `[api]` configuration:

```toml
[api]
base_url = "env:API_BASE_URL"
spec     = "openapi.yaml"   # path to your spec file
```

Tinker supports both YAML and JSON spec formats. The file path is relative to your project root.

### Spec parsing capabilities

The spec parser extracts:

- **Endpoint method, path, summary, and operation ID**
- **Tags** for grouping endpoints
- **Parameterized paths** like `/users/{id}` (matched during `FindEndpoint`)

Endpoints are sorted by path for deterministic output. The parser handles OpenAPI 2.0 (Swagger), 3.0, and 3.1 specs.

## Interactive API Explorer

```bash
tinker api explore
```

Opens an interactive REPL for exploring your API. The explorer uses your OpenAPI spec to provide endpoint discovery and easy invocation.

**Available commands in the explorer:**

| Command | Description |
|---------|-------------|
| `list` or `ls` | List all endpoints (grouped by tag) |
| `tags` | List all available tags |
| `call <method> <path> [body]` | Call an endpoint (e.g., `call GET /users`) |
| `find <keyword>` | Search endpoints by path or summary |
| `quit` or `q` | Exit the explorer |

When you call an endpoint in the explorer, Tinker shows the endpoint's summary from the spec (if available) and automatically uses your configured authentication and headers.

**Example session:**

```
api> tags
  Available tags: users, auth, posts

api> list
  GET    /users          List users
  POST   /users          Create user
  GET    /users/{id}     Get user
  ...

api> call GET /users
  Calling: GET /users (List users)
  { "data": [...], "total": 42 }

api> find login
  POST /auth/login  Login

api> quit
```

## HTTP Session Persistence

API requests automatically persist cookies and auth state across invocations. Session data is stored in `.tinker/session.json` within your project directory.

### How it works

1. When you make an API request, Tinker creates an HTTP client with a cookie jar
2. Any cookies received in the response are stored in the cookie jar
3. On subsequent requests, the stored cookies are sent automatically
4. Auth tokens and custom headers from server responses are also persisted

This means login flows work naturally: authenticate once, then subsequent requests carry the session cookies.

### Example

```bash
# Login — cookies are saved
tinker api POST /auth/login '{"user":"admin","pass":"secret"}'

# Subsequent request — cookies sent automatically
tinker api GET /users

# View session state
tinker api session show

# Clear session when done
tinker api session clear
```

### Session management

```bash
# Show current session state
tinker api session show
```

Displays persisted auth tokens (truncated for security), auth type, and custom headers.

```bash
# Clear all persisted session data
tinker api session clear
```

Removes all cookies, auth tokens, and custom headers from the session store.

### Session storage

Session data is stored in `.tinker/session.json` relative to your project root. This file should be added to your `.gitignore` to avoid committing sensitive session data:

```gitignore
.tinker/
```

## Base URL Resolution

The `base_url` configuration supports flexible formats:

| Config Value | Resolved URL |
|-------------|-------------|
| `http://localhost:8080` | `http://localhost:8080` (used as-is) |
| `https://api.example.com` | `https://api.example.com` (used as-is) |
| `:3000` | `http://localhost:3000` (auto-prefixed) |
| `8080` | `http://localhost:8080` (port-only auto-prefixed) |
| `api.example.com` | `http://api.example.com` (auto-prefixed) |

Trailing slashes are automatically stripped.
