# Plugins

Tinker supports a plugin system that lets you extend its functionality with custom scripts and lifecycle hooks. Plugins can be as simple as a shell script or as sophisticated as a Go program that registers hook handlers.

## Script Plugins

The simplest way to create a plugin is to place an executable script in the `plugins/` directory of your project root. Tinker automatically discovers and registers these scripts as runnable plugins.

### Directory structure

```bash
plugins/
  migrate.sh     # becomes: tinker plugin run migrate
  seed.py        # becomes: tinker plugin run seed
  deploy.bash    # becomes: tinker plugin run deploy
  refresh        # becomes: tinker plugin run refresh (no extension)
```

### Plugin discovery

When you run `tinker plugin list` or `tinker plugin run`, Tinker scans the configured plugins directory (default: `plugins/` in your project root) and loads all executable files as plugins. A file is considered a script plugin if:

- It has the execute permission bit set (on Unix systems), OR
- It has a known script extension (`.sh`, `.bash`, `.zsh`, `.py`, `.rb`, `.js`, `.ts`)

The plugin name is derived from the filename with the extension stripped. For example:
- `migrate.sh` becomes plugin `migrate`
- `seed.py` becomes plugin `seed`
- `deploy.bash` becomes plugin `deploy`
- `refresh` (no extension, but executable) stays as `refresh`

### Running plugins

```bash
# List all loaded plugins
tinker plugin list
# Alias: tinker plugin ls

# Run a plugin
tinker plugin run migrate

# Run a plugin with arguments
tinker plugin run seed --fresh
tinker plugin run deploy production
```

Arguments passed after the plugin name are forwarded directly to the script as command-line arguments. This means your scripts can accept parameters just like any other command-line tool.

### Script examples

**Shell script** (`plugins/migrate.sh`):
```bash
#!/bin/bash
set -e
echo "Running migrations..."
go run ./cmd/migrate "$@"
echo "Done!"
```

**Python script** (`plugins/seed.py`):
```python
#!/usr/bin/env python3
import subprocess
import sys

print("Seeding database...")
subprocess.run(["go", "run", "./cmd/seed"] + sys.argv[1:], check=True)
print("Seed complete!")
```

**Node.js script** (`plugins/generate.js`):
```javascript
#!/usr/bin/env node
const { execSync } = require('child_process');
console.log('Generating code...');
execSync('npm run generate', { stdio: 'inherit' });
console.log('Done!');
```

### Configuration

You can configure the plugin directory and enable/disable the plugin system in `tinker.toml`:

```toml
[plugin]
dir = "plugins"     # directory containing script plugins
enabled = true      # whether the plugin system is active
```

If `enabled = false`, plugins are not loaded and `tinker plugin run` will return an error.

## Hook System

The Go plugin API supports lifecycle hooks — named entry points where plugins can register behavior that runs automatically at specific times during Tinker's execution.

### Available hooks

| Hook | When it fires | Use case |
|------|--------------|----------|
| `pre_db_query` | Before a database query is executed | Log queries, add query hints, validate SQL |
| `post_db_query` | After a database query is executed | Log results, track query performance |
| `pre_api_request` | Before an API request is sent | Add request headers, log requests |
| `post_api_request` | After an API response is received | Log responses, validate status codes |
| `pre_grpc_call` | Before a gRPC call | Log calls, inject metadata |
| `post_grpc_call` | After a gRPC call | Log responses, track latency |
| `init` | During `tinker init` after auto-detection | Add custom detection logic, modify generated config |

### Hook context

Each hook invocation receives a `Context` object that carries data to and from hook handlers:

```go
type Context struct {
    Data map[string]interface{}
}
```

Handlers can read and write to the context, allowing plugins to communicate with each other and modify Tinker's behavior. If a handler returns an error, the hook chain stops and the error is propagated back to the user.

### Go plugin API

For Go-based plugins, use the `plugin` package to register hooks programmatically:

```go
package main

import (
    "fmt"
    "github.com/mvaliolahi/tinker/internal/plugin"
)

func init() {
    p := plugin.New("my-plugin")
    p.Description = "Custom logging and validation"
    p.Version = "1.0.0"

    p.On(plugin.HookPreDBQuery, func(ctx *plugin.Context) error {
        sql := ctx.GetString("sql")
        fmt.Printf("[my-plugin] About to execute: %s\n", sql)
        return nil
    })

    p.On(plugin.HookPostAPIRequest, func(ctx *plugin.Context) error {
        statusCode := ctx.Get("status_code")
        fmt.Printf("[my-plugin] Response status: %v\n", statusCode)
        return nil
    })

    plugin.Global().Register(p)
}
```

### Hook dispatch order

Hooks are dispatched to all registered plugins in **plugin registration order** (sorted alphabetically by plugin name). If any handler returns an error, the dispatch stops immediately and the error is returned. This ensures that a failing plugin doesn't silently corrupt the hook chain.

## Plugin Registry

The global `Registry` manages all loaded plugins:

- `plugin.Global()` returns the global registry singleton
- `Register(p)` adds a plugin (fails if a plugin with the same name already exists)
- `Unregister(name)` removes a plugin
- `Get(name)` retrieves a plugin by name
- `List()` returns all plugin names, sorted alphabetically
- `ListPlugins()` returns detailed info for all plugins (name, description, version, hooks)
- `Dispatch(hook, ctx)` sends a hook event to all registered handlers

## Best Practices

1. **Make scripts executable** — On Unix, run `chmod +x plugins/your-script.sh` to ensure Tinker can detect and run your plugin
2. **Use shebangs** — Always include a `#!/bin/bash` or `#!/usr/bin/env python3` line so the system knows how to execute your script
3. **Return proper exit codes** — Exit with code 0 on success and non-zero on failure, so Tinker can report errors correctly
4. **Keep plugins focused** — Each plugin should do one thing well. If you need multiple behaviors, create separate scripts
5. **Add to .gitignore carefully** — Plugin scripts are typically committed to version control so your team shares the same workflow. However, if a plugin contains sensitive information (e.g., deployment keys), add it to `.gitignore`
6. **Use hooks sparingly** — Hook handlers run synchronously, so slow handlers will delay the operation. Keep hook logic fast and lightweight
