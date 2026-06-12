package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Hook represents a named entry point where plugins can register behavior.
type Hook string

const (
	// HookPreDBQuery runs before a database query is executed.
	HookPreDBQuery Hook = "pre_db_query"
	// HookPostDBQuery runs after a database query is executed.
	HookPostDBQuery Hook = "post_db_query"
	// HookPreAPIRequest runs before an API request is sent.
	HookPreAPIRequest Hook = "pre_api_request"
	// HookPostAPIRequest runs after an API response is received.
	HookPostAPIRequest Hook = "post_api_request"
	// HookPreGRPCCall runs before a gRPC call.
	HookPreGRPCCall Hook = "pre_grpc_call"
	// HookPostGRPCCall runs after a gRPC call.
	HookPostGRPCCall Hook = "post_grpc_call"
	// HookInit runs during tinker init after auto-detection.
	HookInit Hook = "init"
)

// Context carries data to and from hook handlers.
type Context struct {
	Data map[string]interface{}
}

// NewContext creates an empty hook context.
func NewContext() *Context {
	return &Context{Data: make(map[string]interface{})}
}

// Set stores a value in the context.
func (c *Context) Set(key string, val interface{}) {
	c.Data[key] = val
}

// Get retrieves a value from the context.
func (c *Context) Get(key string) interface{} {
	return c.Data[key]
}

// GetString retrieves a string value from the context.
func (c *Context) GetString(key string) string {
	if v, ok := c.Data[key].(string); ok {
		return v
	}
	return ""
}

// Handler is a function that handles a hook invocation.
// It can modify the context and optionally return an error to abort the chain.
type Handler func(*Context) error

// Plugin represents a loaded plugin with its metadata and registered hooks.
type Plugin struct {
	Name        string
	Description string
	Version     string
	Meta        map[string]interface{}
	handlers    map[Hook][]Handler
}

// New creates a new plugin with the given name.
func New(name string) *Plugin {
	return &Plugin{
		Name:     name,
		Meta:     make(map[string]interface{}),
		handlers: make(map[Hook][]Handler),
	}
}

// On registers a handler for a specific hook.
func (p *Plugin) On(hook Hook, handler Handler) {
	p.handlers[hook] = append(p.handlers[hook], handler)
}

// Registry manages all loaded plugins and dispatches hooks.
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]*Plugin
}

// Global registry instance.
var global = NewRegistry()

// Global returns the global plugin registry.
func Global() *Registry {
	return global
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p *Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[p.Name]; exists {
		return fmt.Errorf("plugin %q already registered", p.Name)
	}
	r.plugins[p.Name] = p
	return nil
}

// Unregister removes a plugin from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.plugins, name)
}

// Get returns a plugin by name.
func (r *Registry) Get(name string) (*Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

// List returns all registered plugin names, sorted.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListPlugins returns all registered plugins with their info.
func (r *Registry) ListPlugins() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []Info
	for _, p := range r.plugins {
		result = append(result, Info{
			Name:        p.Name,
			Description: p.Description,
			Version:     p.Version,
			Hooks:       p.registeredHooks(),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Info provides metadata about a loaded plugin.
type Info struct {
	Name        string
	Description string
	Version     string
	Hooks       []string
}

// Dispatch sends a hook event to all registered handlers across all plugins.
// Handlers are called in plugin registration order. If any handler returns an
// error, dispatch stops and returns that error.
func (r *Registry) Dispatch(hook Hook, ctx *Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, name := range r.sortedNames() {
		p := r.plugins[name]
		if handlers, ok := p.handlers[hook]; ok {
			for _, h := range handlers {
				if err := h(ctx); err != nil {
					return fmt.Errorf("plugin %q hook %s: %w", name, hook, err)
				}
			}
		}
	}
	return nil
}

func (r *Registry) sortedNames() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p *Plugin) registeredHooks() []string {
	var hooks []string
	for h := range p.handlers {
		hooks = append(hooks, string(h))
	}
	sort.Strings(hooks)
	return hooks
}

// LoadScripts loads all executable scripts from a directory as command plugins.
// Scripts are registered as custom commands that can be invoked via `tinker plugin run <name>`.
func LoadScripts(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading scripts directory: %w", err)
	}

	var loaded int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		path := filepath.Join(dir, name)
		info, err := e.Info()
		if err != nil {
			continue
		}

		// On Unix, check if executable; on Windows, check common script extensions
		if info.Mode()&0111 != 0 || isScriptExtension(name) {
			scriptName := cleanScriptName(name)
			p := New("script:" + scriptName)
			p.Description = "Script: " + name
			p.Meta["script_path"] = path
			p.Meta["original_name"] = name
			if err := Global().Register(p); err != nil {
				continue // skip duplicates
			}
			loaded++
		}
	}
	return loaded, nil
}

// RunScript executes a registered script plugin.
func RunScript(name string, args []string) error {
	p, ok := Global().Get("script:" + name)
	if !ok {
		return fmt.Errorf("script plugin %q not found", name)
	}

	scriptPath, _ := p.Meta["script_path"].(string)
	if scriptPath == "" {
		return fmt.Errorf("script plugin %q has no script path", name)
	}

	cmd := exec.Command(scriptPath, args...) //nolint:gosec // script path from our own plugin registry
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

// ListScripts returns the names of all loaded script plugins (without the "script:" prefix).
func ListScripts() []string {
	var scripts []string
	for _, name := range Global().List() {
		if strings.HasPrefix(name, "script:") {
			scripts = append(scripts, strings.TrimPrefix(name, "script:"))
		}
	}
	return scripts
}

// cleanScriptName strips extensions from script names for a cleaner alias.
func cleanScriptName(name string) string {
	for _, ext := range []string{".sh", ".bash", ".zsh", ".py", ".rb", ".js", ".ts"} {
		if strings.HasSuffix(name, ext) {
			return strings.TrimSuffix(name, ext)
		}
	}
	return name
}

// isScriptExtension checks if a filename has a known script extension.
func isScriptExtension(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range []string{".sh", ".bash", ".zsh", ".py", ".rb", ".js", ".ts"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
