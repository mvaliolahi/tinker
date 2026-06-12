# gRPC

Tinker provides native gRPC support for listing services, describing service methods, and invoking gRPC methods. The `list` and `describe` commands work natively via server reflection — no external binary required.

## Quick Reference

```bash
tinker grpc list                                    # List services (native)
tinker grpc describe UserService                    # Describe a service (native)
tinker grpc call UserService/GetUser '{"id": 1}'   # Call a method (requires grpcurl)
tinker grpc                                        # Interactive REPL (requires evans)
```

## Configuration

Add a `[grpc]` section to your `tinker.toml`:

```toml
[grpc]
addr       = "env:GRPC_ADDR"     # gRPC server address (e.g., "localhost:50051")
proto_dir  = "./proto"           # Directory containing .proto files
reflection = true                # Whether server reflection is enabled
```

| Field | Required | Description |
|-------|----------|-------------|
| `addr` | Yes* | gRPC server address source. Supports `env:VAR_NAME` or direct address |
| `proto_dir` | No | Path to `.proto` files. Used for proto-file mode when reflection is disabled |
| `reflection` | No | Enable server reflection. Required for native `list` and `describe` commands. Default: `false` |

*Either `addr` or `proto_dir` must be set.

### Environment variable

```bash
# .env
GRPC_ADDR=localhost:50051
```

```toml
# tinker.toml
[grpc]
addr       = "env:GRPC_ADDR"
reflection = true
```

## Commands

### List Services

```bash
tinker grpc list
```

Lists all gRPC services available on the server. This command uses **native server reflection** — no `grpcurl` binary is required. The connection is established using an insecure transport (suitable for local development), with a 10-second dial timeout.

Example output:

```
  GRPC  Services

  grpc.reflection.v1alpha.ServerReflection
  UserService
  OrderService
  HealthCheck
```

> **Note:** Server reflection must be enabled on the gRPC server and `reflection = true` must be set in your `tinker.toml`. If reflection is disabled, you need `grpcurl` with proto files instead.

### Describe Service

```bash
tinker grpc describe <service>
```

Shows the method signatures for a gRPC service, including method names and request/response types. This also uses native server reflection — no external tools needed.

```bash
tinker grpc describe UserService
```

Example output:

```
  GRPC  Describe UserService

  service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}
    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {}
    rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {}
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {}
  }
```

### Call Method

```bash
tinker grpc call <method> [data]
```

Invokes a gRPC method with JSON request data. This command requires the `grpcurl` CLI tool for proper proto serialization, as native Go proto encoding requires the full proto descriptor at runtime.

```bash
tinker grpc call UserService/GetUser '{"id": 1}'
tinker grpc call UserService/CreateUser '{"name": "Ali", "email": "ali@example.com"}'
```

If `grpcurl` is not installed, Tinker will display an error with installation instructions. Install it with:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Interactive Session

```bash
tinker grpc
```

Opens an interactive gRPC REPL using `evans`. This provides a rich interactive experience with:

- Service and method listing
- Request building with field completion
- Response inspection
- Header and trailer viewing

`evans` must be installed separately:

```bash
go install github.com/ktr0731/evans@latest
```

## Native gRPC Architecture

Tinker's gRPC module uses a two-tier approach similar to the database module:

### Native client (zero dependencies for list/describe)

The `NativeClient` in `internal/grpc/native.go` uses the `google.golang.org/grpc` library directly:

1. **Connection** — Establishes a gRPC connection with insecure credentials and a 10-second block timeout
2. **Server reflection** — Uses `grpc.reflection.v1alpha.ServerReflectionClient` to query service information
3. **File descriptor parsing** — Implements a minimal protobuf wire format parser to extract service names and method signatures from file descriptors returned by reflection

This means `list` and `describe` work without any external tools, making them ideal for quick service discovery during development.

### CLI fallback (grpcurl for call, evans for REPL)

For the `call` command and interactive REPL, Tinker shells out to external tools:

| Command | Tool | Why it's needed |
|---------|------|-----------------|
| `call` | `grpcurl` | Dynamic proto serialization for request encoding |
| Interactive | `evans` | Full-featured REPL with field completion and inspection |

### When to use native vs. CLI

| Use Case | Native | CLI |
|----------|--------|-----|
| List services | Yes | Also works |
| Describe a service | Yes | Also works |
| Call a method | No | `grpcurl` required |
| Interactive session | No | `evans` required |
| Server without reflection | No | `grpcurl` with `--proto` files |

## Auto-Detection

When you run `tinker init`, Tinker scans for:

1. **GRPC_ADDR environment variable** — If `GRPC_ADDR` is found in `.env`, it's configured as the gRPC server address
2. **Proto directories** — `proto/`, `protos/`, `api/proto/` directories containing `.proto` files
3. **Docker Compose services** — Services with gRPC-related image names (e.g., `grpc`, `grpcurl`) are auto-detected

The detection results are reflected in the generated `tinker.toml`:

```toml
[grpc]
addr       = "env:GRPC_ADDR"
proto_dir  = "./proto"
reflection = true
```
