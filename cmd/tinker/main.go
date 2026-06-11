package main

import (
	"fmt"
	"os"

	"github.com/mvaliolahi/tinker/internal/api"
	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/contract"
	"github.com/mvaliolahi/tinker/internal/db"
	"github.com/mvaliolahi/tinker/internal/detect"
	"github.com/mvaliolahi/tinker/internal/grpc"
	"github.com/mvaliolahi/tinker/internal/run"
	"github.com/spf13/cobra"
)

var (
	projectDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinker",
		Short: "A project-aware CLI for database, API, and gRPC interaction",
		Long: `Tinker is a project-aware CLI that composes best-in-class open source tools
(usql, httpie, grpcurl, evans) into a unified interface for working with
your project's database, HTTP APIs, and gRPC services.

It reads your project's tinker.toml (and .env) to automatically configure
connections, so you can jump straight into working with your project.

Quick start:
  tinker init          # Auto-detect project config and generate tinker.toml
  tinker db            # Open interactive database session
  tinker api GET /users  # Call an API endpoint
  tinker grpc list     # List gRPC services`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: auto-detect from current dir)")

	// Add subcommands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(dbCmd())
	rootCmd.AddCommand(apiCmd())
	rootCmd.AddCommand(grpcCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// loadConfig loads the tinker configuration, finding the project root if needed.
func loadConfig() (*config.Config, string, error) {
	dir := projectDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", fmt.Errorf("getting current directory: %w", err)
		}
		dir = cwd
	}

	// Find project root with tinker.toml
	root, err := config.FindProjectRoot(dir)
	if err != nil {
		return nil, "", fmt.Errorf("could not find tinker.toml. Run 'tinker init' first: %w", err)
	}

	cfg, err := config.Load(root)
	if err != nil {
		return nil, "", fmt.Errorf("loading config: %w", err)
	}

	return cfg, root, nil
}

// initCmd creates the `tinker init` command.
func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Auto-detect project configuration and generate tinker.toml",
		Long: `Scans your project directory for common patterns like .env files,
OpenAPI specs, proto files, and database configurations, then generates
a tinker.toml file with sensible defaults.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := projectDir
			if dir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				dir = cwd
			}

			// Check if tinker.toml already exists
			tinkerPath := fmt.Sprintf("%s/tinker.toml", dir)
			if _, err := os.Stat(tinkerPath); err == nil {
				fmt.Fprintf(os.Stderr, "tinker.toml already exists at %s\n", tinkerPath)
				fmt.Fprintf(os.Stderr, "Remove it first if you want to regenerate, or edit it manually.\n")
				return nil
			}

			fmt.Printf("Scanning project at %s...\n\n", dir)

			detector := detect.NewDetector(dir)
			result, err := detector.Detect()
			if err != nil {
				return fmt.Errorf("detecting project config: %w", err)
			}

			// Print what was found
			if result.Database != nil {
				fmt.Printf("  ✓ Database detected: %s (source: %s)\n", result.Database.Type, result.Database.Source)
			}
			if result.API != nil {
				fmt.Printf("  ✓ API detected")
				if result.API.BaseURL != "" {
					fmt.Printf(" (base_url: %s)", result.API.BaseURL)
				}
				if result.API.Spec != "" {
					fmt.Printf(" (spec: %s)", result.API.Spec)
				}
				fmt.Println()
			}
			if result.GRPC != nil {
				fmt.Printf("  ✓ gRPC detected")
				if result.GRPC.Addr != "" {
					fmt.Printf(" (addr: %s)", result.GRPC.Addr)
				}
				if result.GRPC.ProtoDir != "" {
					fmt.Printf(" (proto_dir: %s)", result.GRPC.ProtoDir)
				}
				fmt.Println()
			}

			if result.Database == nil && result.API == nil && result.GRPC == nil {
				fmt.Println("  ✗ No database, API, or gRPC configuration detected.")
				fmt.Println()
				fmt.Println("Create a tinker.toml manually. Example:")
				fmt.Println(`
[database]
source = "env:DATABASE_URL"
type = "postgres"

[api]
base_url = "env:API_BASE_URL"
spec = "openapi.yaml"
auth = "env:API_TOKEN"
auth_type = "bearer"

[grpc]
addr = "env:GRPC_ADDR"
proto_dir = "./proto"
reflection = true
`)
				return nil
			}

			if err := contract.GenerateTinkerToml(result, dir); err != nil {
				return fmt.Errorf("generating tinker.toml: %w", err)
			}

			fmt.Printf("\n✓ Generated tinker.toml at %s\n", tinkerPath)
			fmt.Println("  Review and adjust the configuration as needed.")
			return nil
		},
	}

	return cmd
}

// dbCmd creates the `tinker db` command.
func dbCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Interact with your project's database",
		Long: `Opens an interactive database session using usql with connection
details from tinker.toml. Supports subcommands for quick queries.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "connect",
		Short: "Open interactive database session",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			fmt.Printf("Connecting to %s database...\n", session.Type)
			return session.Connect()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "tables",
		Short: "List all tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			output, err := session.Tables()
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "describe [table]",
		Short: "Describe a table's schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			output, err := session.Describe(args[0])
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "count [table]",
		Short: "Count rows in a table",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			where := ""
			if len(args) > 1 {
				where = args[1]
			}

			output, err := session.Count(args[0], where)
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "find [table] [id]",
		Short: "Find a row by ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			output, err := session.Find(args[0], args[1])
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "exec [sql]",
		Short: "Execute a SQL statement",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := db.NewSession(cfg.Database)
			if err != nil {
				return err
			}

			output, err := session.Exec(args[0])
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	// Default: open interactive session
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, _, err := loadConfig()
		if err != nil {
			return err
		}

		session, err := db.NewSession(cfg.Database)
		if err != nil {
			return err
		}

		fmt.Printf("Connecting to %s database...\n", session.Type)
		return session.Connect()
	}

	return cmd
}

// apiCmd creates the `tinker api` command.
func apiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api [method] [path] [data]",
		Short: "Call your project's HTTP API endpoints",
		Long: `Makes HTTP requests to your project's API with automatic
authentication and base URL resolution from tinker.toml.

Examples:
  tinker api GET /users
  tinker api POST /users '{"name": "Ali"}'
  tinker api PUT /users/1 '{"name": "Updated"}'
  tinker api DELETE /users/1`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "interactive",
		Short: "Open interactive HTTP session (requires httpie or curlie)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := api.NewSession(cfg.API)
			if err != nil {
				return err
			}

			return session.Interactive()
		},
	})

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, _, err := loadConfig()
		if err != nil {
			return err
		}

		session, err := api.NewSession(cfg.API)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return session.Interactive()
		}

		method := "GET"
		path := "/"
		var body string

		switch len(args) {
		case 1:
			path = args[0]
		case 2:
			method = args[0]
			path = args[1]
		case 3:
			method = args[0]
			path = args[1]
			body = args[2]
		default:
			return fmt.Errorf("too many arguments. Usage: tinker api [METHOD] [PATH] [BODY]")
		}

		output, err := session.Request(method, path, body, nil)
		if err != nil {
			return err
		}
		fmt.Print(output)
		return nil
	}

	return cmd
}

// grpcCmd creates the `tinker grpc` command.
func grpcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grpc",
		Short: "Interact with your project's gRPC services",
		Long: `Calls gRPC services using grpcurl/evans with connection
details from tinker.toml. Supports server reflection and proto file loading.

Examples:
  tinker grpc                    # Open interactive REPL (evans)
  tinker grpc list               # List available services
  tinker grpc describe UserService  # Describe a service
  tinker grpc call UserService/GetUser '{"id": 1}'`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available gRPC services",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := grpc.NewSession(cfg.GRPC)
			if err != nil {
				return err
			}

			output, err := session.ListServices()
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "describe [service]",
		Short: "Describe a gRPC service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := grpc.NewSession(cfg.GRPC)
			if err != nil {
				return err
			}

			output, err := session.DescribeService(args[0])
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "call [method] [data]",
		Short: "Call a gRPC method",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadConfig()
			if err != nil {
				return err
			}

			session, err := grpc.NewSession(cfg.GRPC)
			if err != nil {
				return err
			}

			data := ""
			if len(args) > 1 {
				data = args[1]
			}

			output, err := session.Call(args[0], data)
			if err != nil {
				return err
			}
			fmt.Print(output)
			return nil
		},
	})

	// Default: open interactive REPL
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, _, err := loadConfig()
		if err != nil {
			return err
		}

		session, err := grpc.NewSession(cfg.GRPC)
		if err != nil {
			return err
		}

		return session.Interactive()
	}

	return cmd
}

// runCmd creates the `tinker run` command.
func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [code]",
		Short: "Execute one-off Go code within your project context",
		Long: `Generates a temporary Go program that imports your project's
packages, compiles and runs it. Useful for one-off operations like
creating test data or running migrations.

Examples:
  tinker run 'fmt.Println("Hello from tinker!")'
  tinker run '// import "myproject/models"\nmodels.CreateUser(ctx, "Ali")'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, projectRoot, err := loadConfig()
			_ = cfg
			if err != nil {
				// run doesn't strictly need tinker.toml, but it helps
				cwd, _ := os.Getwd()
				projectRoot = cwd
			}

			code, imports := run.ParseCode(args[0])

			runner := run.NewRunner(run.Config{
				ProjectDir: projectRoot,
				Code:       code,
				Imports:    imports,
			})

			return runner.Run()
		},
	}

	return cmd
}

// versionCmd creates the `tinker version` command.
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print tinker version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("tinker v0.1.0")
		},
	}
}
