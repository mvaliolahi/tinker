package plugin

import (
	"fmt"
	"testing"
)

func TestRegistry_RegisterAndList(t *testing.T) {
	r := NewRegistry()

	p := New("test-plugin")
	p.Description = "A test plugin"
	p.Version = "1.0.0"

	if err := r.Register(p); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	names := r.List()
	if len(names) != 1 || names[0] != "test-plugin" {
		t.Errorf("List() = %v, want [test-plugin]", names)
	}

	// Duplicate registration should fail
	if err := r.Register(New("test-plugin")); err == nil {
		t.Error("expected error on duplicate registration")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := NewRegistry()
	r.Register(New("a"))
	r.Register(New("b"))

	r.Unregister("a")

	names := r.List()
	if len(names) != 1 || names[0] != "b" {
		t.Errorf("List() after Unregister = %v, want [b]", names)
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	p := New("my-plugin")
	r.Register(p)

	got, ok := r.Get("my-plugin")
	if !ok {
		t.Error("Get returned not ok")
	}
	if got.Name != "my-plugin" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "my-plugin")
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Get for nonexistent should return false")
	}
}

func TestPlugin_On(t *testing.T) {
	p := New("hook-test")
	called := false
	p.On(HookPreDBQuery, func(_ *Context) error {
		called = true
		return nil
	})

	if len(p.handlers[HookPreDBQuery]) != 1 {
		t.Error("handler not registered")
	}

	ctx := NewContext()
	p.handlers[HookPreDBQuery][0](ctx)

	if !called {
		t.Error("handler was not called")
	}
}

func TestRegistry_Dispatch(t *testing.T) {
	r := NewRegistry()

	var order []string
	p1 := New("alpha")
	p1.On(HookInit, func(_ *Context) error {
		order = append(order, "alpha")
		return nil
	})

	p2 := New("beta")
	p2.On(HookInit, func(_ *Context) error {
		order = append(order, "beta")
		return nil
	})

	r.Register(p1)
	r.Register(p2)

	ctx := NewContext()
	if err := r.Dispatch(HookInit, ctx); err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	if len(order) != 2 || order[0] != "alpha" || order[1] != "beta" {
		t.Errorf("dispatch order = %v, want [alpha beta]", order)
	}
}

func TestRegistry_Dispatch_AbortOnError(t *testing.T) {
	r := NewRegistry()

	p := New("failer")
	p.On(HookPreDBQuery, func(_ *Context) error {
		return fmt.Errorf("boom")
	})
	r.Register(p)

	err := r.Dispatch(HookPreDBQuery, NewContext())
	if err == nil {
		t.Error("expected error from Dispatch")
	}
}

func TestContext_SetGet(t *testing.T) {
	ctx := NewContext()
	ctx.Set("key", "value")

	if v := ctx.GetString("key"); v != "value" {
		t.Errorf("GetString = %q, want %q", v, "value")
	}
	if v := ctx.GetString("missing"); v != "" {
		t.Errorf("GetString missing = %q, want empty", v)
	}
}

func TestListPlugins(t *testing.T) {
	r := NewRegistry()
	p := New("demo")
	p.Description = "Demo plugin"
	p.Version = "2.0"
	p.On(HookInit, func(_ *Context) error { return nil })
	p.On(HookPreDBQuery, func(_ *Context) error { return nil })
	r.Register(p)

	infos := r.ListPlugins()
	if len(infos) != 1 {
		t.Fatalf("ListPlugins returned %d items, want 1", len(infos))
	}
	info := infos[0]
	if info.Name != "demo" {
		t.Errorf("Name = %q, want %q", info.Name, "demo")
	}
	if len(info.Hooks) != 2 {
		t.Errorf("Hooks = %v, want 2 hooks", info.Hooks)
	}
}

func TestCleanScriptName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"migrate.sh", "migrate"},
		{"seed.py", "seed"},
		{"deploy.bash", "deploy"},
		{"noext", "noext"},
		{"readme.md", "readme.md"},
	}
	for _, tt := range tests {
		got := cleanScriptName(tt.input)
		if got != tt.want {
			t.Errorf("cleanScriptName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsScriptExtension(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"run.sh", true},
		{"script.py", true},
		{"deploy.js", true},
		{"readme.md", false},
		{"Makefile", false},
	}
	for _, tt := range tests {
		got := isScriptExtension(tt.input)
		if got != tt.want {
			t.Errorf("isScriptExtension(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
