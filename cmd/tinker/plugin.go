package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mvaliolahi/tinker/internal/plugin"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func pluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"},
		Short:   "Manage and run plugins",
	}

	cmd.AddCommand(
		pluginListCmd(),
		pluginRunCmd(),
	)

	return cmd
}

func pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List loaded plugins",
		RunE: func(_ *cobra.Command, _ []string) error {
			// Load plugins from project
			loadProjectPlugins()

			infos := plugin.Global().ListPlugins()
			if len(infos) == 0 {
				fmt.Println(ui.Dim("  No plugins loaded"))
				fmt.Println(ui.Hint("Add executable scripts to a plugins/ directory in your project"))
				return nil
			}

			fmt.Println()
			fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Plugins"))
			fmt.Println()

			for _, info := range infos {
				desc := ""
				if info.Description != "" {
					desc = ui.Dim(" — " + info.Description)
				}
				fmt.Printf("  %-20s%s\n", ui.Accent(info.Name), desc)
				if info.Version != "" {
					fmt.Println(ui.KeyValue("version", info.Version))
				}
				if len(info.Hooks) > 0 {
					fmt.Println(ui.KeyValue("hooks", fmt.Sprintf("%v", info.Hooks)))
				}
			}

			fmt.Println()
			return nil
		},
	}
}

func pluginRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <name> [args...]",
		Short: "Run a script plugin",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			loadProjectPlugins()

			name := args[0]
			scriptArgs := args[1:]

			fmt.Println("  " + ui.CfgLabel() + " " + ui.Bold("Plugin: "+name))
			fmt.Println()

			if err := plugin.RunScript(name, scriptArgs); err != nil {
				return fmt.Errorf("plugin %q: %w", name, err)
			}
			return nil
		},
	}
}

// loadProjectPlugins loads script plugins from the project's plugins/ directory.
func loadProjectPlugins() {
	dir, err := resolveDir()
	if err != nil {
		return
	}

	pluginsDir := filepath.Join(dir, "plugins")
	if info, err := os.Stat(pluginsDir); err == nil && info.IsDir() {
		_, _ = plugin.LoadScripts(pluginsDir)
	}
}
