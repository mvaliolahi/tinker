package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func commandsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cmd",
		Aliases: []string{"commands"},
		Short:   "Run custom project commands from tinker.toml",
	}

	cmd.AddCommand(commandsListCmd())

	// The main RunE handles: tinker cmd <command-name> [args...]
	cmd.RunE = func(_ *cobra.Command, args []string) error {
		cfg, _, err := loadConfig()
		if err != nil {
			return err
		}

		if len(cfg.Commands) == 0 {
			fmt.Println(ui.Warning("No custom commands defined"))
			fmt.Println(ui.Hint("Add a [commands] section to tinker.toml"))
			fmt.Println()
			fmt.Println(ui.Dim("  Example:"))
			fmt.Println(ui.Dim("    [commands]"))
			fmt.Println(ui.Dim("      migrate = \"go run ./cmd/migrate\""))
			fmt.Println(ui.Dim("      seed    = \"go run ./cmd/seed\""))
			fmt.Println(ui.Dim("      test    = \"go test ./...\""))
			return nil
		}

		if len(args) == 0 {
			// No args — list available commands
			return listCommands(cfg.Commands)
		}

		name := args[0]
		commandStr, ok := cfg.Commands[name]
		if !ok {
			fmt.Println(ui.Error("Command not found: " + name))
			fmt.Println(ui.Hint("Use 'tinker cmd' to list available commands"))
			return nil
		}

		// Append any extra args to the command
		if len(args) > 1 {
			commandStr += " " + strings.Join(args[1:], " ")
		}

		fmt.Println("  " + ui.MakeLabel() + " " + ui.Bold(name))
		fmt.Println(ui.KeyValue("command", commandStr))
		fmt.Println()

		return runCustomCommand(commandStr)
	}

	return cmd
}

func commandsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List custom commands",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			return listCommands(cfg.Commands)
		},
	}
}

func listCommands(commands map[string]string) error {
	if len(commands) == 0 {
		fmt.Println(ui.Warning("No custom commands defined"))
		fmt.Println(ui.Hint("Add a [commands] section to tinker.toml"))
		return nil
	}

	fmt.Println()
	fmt.Println("  " + ui.MakeLabel() + " " + ui.Bold("Commands"))
	fmt.Println()

	// Sort command names for deterministic output
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Println(ui.Bullet(name, ui.Dim(commands[name])))
	}

	fmt.Println()
	fmt.Println(ui.Hint("Use: tinker cmd <name> [args...]"))
	return nil
}

// runCustomCommand executes a custom command string via the shell.
func runCustomCommand(commandStr string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	var cmd *exec.Cmd
	if strings.Contains(shell, "zsh") || strings.Contains(shell, "bash") {
		cmd = exec.Command(shell, "-c", commandStr)
	} else {
		cmd = exec.Command("/bin/sh", "-c", commandStr)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
