# Docker Compose Integration

Tinker can detect and inspect Docker Compose services in your project. This is useful for understanding your project's infrastructure, identifying service types (database, API, gRPC), and verifying that your Docker Compose configuration aligns with your Tinker configuration.

## Auto-Detection

When you run `tinker init`, Tinker scans for Docker Compose files in your project root. The following filenames are recognized:

- `docker-compose.yml`
- `docker-compose.yaml`
- `compose.yml`
- `compose.yaml`

Additionally, environment-specific Compose files are detected:

- `docker-compose.staging.yml`
- `docker-compose.production.yml`
- `compose.staging.yml`
- `compose.production.yml`

If a Docker Compose file is found, Tinker:
1. Parses the YAML to extract service definitions
2. Auto-detects service types using heuristics (image names, port numbers, environment variables)
3. Includes Docker information in the generated `tinker.toml`

## Service Type Detection

Tinker uses heuristics to automatically classify services as database, API, or gRPC:

### Database detection

A service is classified as a database if its image name contains any of:
- `postgres`, `mysql`, `mariadb`, `mongo`, `redis`, `sqlite`, `cockroach`, `cassandra`, `couchdb`, `dynamodb`, `elasticsearch`

Or if it exposes well-known database ports:
- 5432 (PostgreSQL), 3306 (MySQL), 27017 (MongoDB), 6379 (Redis), etc.

### API detection

A service is classified as an API if its image name contains:
- `nginx`, `caddy`, `traefik`, `api`, `server`, `web`, `express`, `fastapi`, `flask`, `rails`, `django`

### gRPC detection

A service is classified as gRPC if its image name contains:
- `grpc`, `grpcurl`, `evans`

Or if it exposes gRPC-typical ports:
- 50051, 9090

## Commands

### List services

```bash
tinker docker list
# Alias: tinker docker ls
```

Lists all Docker Compose services with their detected types:

```
  Docker Compose Services
  file: docker-compose.yml

  db          postgres:15        ports: 5432:5432
    detected: DB database
  api         myapp:latest       ports: 8080:8080
    detected: API api
  grpc-server grpc-service:latest  ports: 50051:50051
    detected: GRPC grpc
  redis       redis:7            ports: 6379:6379
```

### Service details

```bash
tinker docker info
```

Shows detailed information about each service:

```
  Docker Compose
  file: docker-compose.yml
  services: 4

  db
    image: postgres:15
    ports: 5432:5432
    type: [database]

  api
    image: myapp:latest
    ports: 8080:8080
    type: [api]
```

### Without subcommand

```bash
tinker docker
```

Shows a summary of Docker Compose services. If no Docker Compose file is found, displays a hint about adding one.

## Dashboard Integration

When you run `tinker` without arguments (the dashboard), Docker Compose information is displayed if a compose file is detected. The dashboard shows:

- Whether Docker Compose is present
- Number of services
- Service types (database, API, gRPC)

This gives you a quick visual overview of your project's containerized infrastructure alongside your database, API, and gRPC configuration.

## Configuration in tinker.toml

Docker Compose information is primarily auto-detected and displayed, not explicitly configured in `tinker.toml`. However, the contract generator adds comments about detected compose files:

```toml
# Auto-detected Docker Compose file: docker-compose.yml
# Services: db (postgres), api (nginx), grpc-server (grpc)
```

If you change your Docker Compose configuration, the changes will be reflected the next time you run `tinker docker list` or `tinker init`. There's no need to manually update `tinker.toml` for Docker-related changes.

## Environment-Specific Compose Files

Tinker also detects environment-specific Compose files. When you use `tinker --env staging`, Tinker looks for:

- `docker-compose.staging.yml`
- `compose.staging.yml`

These are checked in addition to the base Compose file. The detected services are combined from both files.

## Working with Docker and Tinker

Here are common workflows combining Docker and Tinker:

### Start services, then interact

```bash
# Start your Docker services
docker compose up -d

# Wait for services to be ready
tinker db ping

# Interact with the database
tinker db tables
tinker db migrate up
tinker db seed

# Test the API
tinker api GET /health
```

### Check service status

```bash
# See what Docker services are configured
tinker docker list

# Check database connectivity
tinker db ping

# Check API availability
tinker api GET /health
```

### Multi-environment with Docker

```bash
# Development (default)
tinker db tables

# Staging
tinker --env staging db tables

# Production
tinker --env production db ping
```
