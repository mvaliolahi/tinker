package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mvaliolahi/tinker/internal/api"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func apiCmd() *cobra.Command {
	var jqFilter string

	cmd := &cobra.Command{
		Use:   "api [method] [path] [body]",
		Short: "Call your project's HTTP API",
	}

	cmd.Flags().StringVarP(&jqFilter, "jq", "q", "", "Filter response with gjson path or jq expression")

	cmd.AddCommand(
		apiEndpointsCmd(),
		apiExploreCmd(),
		apiSessionCmd(),
	)

	cmd.RunE = func(_ *cobra.Command, args []string) error {
		opts, err := apiSessionOpts(jqFilter)
		if err != nil {
			return err
		}

		s, err := newAPISessionWithOpts(opts...)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			fmt.Println("  " + ui.APILabel() + " " + ui.Bold("Interactive session"))
			fmt.Println(ui.KeyValue("base", s.BaseURL))
			fmt.Println()
			return s.Interactive()
		}

		method, path, body := parseAPICall(args)
		fmt.Println()
		fmt.Println("  " + ui.APILabel() + " " + ui.HTTPMethod(method) + " " + ui.Accent(path))
		fmt.Println(ui.KeyValue("url", s.BaseURL+path))
		if body != "" {
			fmt.Println(ui.KeyValue("body", body))
		}
		if jqFilter != "" {
			fmt.Println(ui.KeyValue("jq", jqFilter))
		}
		fmt.Println()

		out, err := s.Request(method, path, body, nil)
		fmt.Print(out)
		return err
	}

	return cmd
}

// apiEndpointsCmd lists all endpoints from the OpenAPI spec defined in tinker.toml.
func apiEndpointsCmd() *cobra.Command {
	var tagFilter string

	cmd := &cobra.Command{
		Use:     "endpoints",
		Aliases: []string{"ep"},
		Short:   "List API endpoints from OpenAPI spec",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}
			if cfg.API == nil || cfg.API.Spec == "" {
				fmt.Println(ui.Warning("No OpenAPI spec configured"))
				fmt.Println(ui.Hint("Add spec = \"openapi.yaml\" to the [api] section in tinker.toml"))
				return nil
			}

			specPath := cfg.API.Spec
			if !filepath.IsAbs(specPath) {
				specPath = filepath.Join(root, specPath)
			}

			spec, err := api.ParseSpec(specPath)
			if err != nil {
				return fmt.Errorf("parsing spec: %w", err)
			}

			fmt.Println()
			fmt.Println("  " + ui.APILabel() + " " + ui.Bold("Endpoints"))
			if spec.Title != "" {
				fmt.Println(ui.KeyValue("spec", spec.Title+" v"+spec.Version))
			}
			fmt.Println(ui.KeyValue("total", fmt.Sprintf("%d endpoint(s)", len(spec.Endpoints))))
			fmt.Println()

			if len(spec.Endpoints) == 0 {
				fmt.Println(ui.Dim("  No endpoints found in spec"))
				return nil
			}

			// Group by tag if filter is set, otherwise list all
			if tagFilter != "" {
				byTag := spec.EndpointsByTag()
				endpoints, ok := byTag[tagFilter]
				if !ok {
					fmt.Println(ui.Warning("Tag not found: " + tagFilter))
					availableTags := make([]string, 0, len(byTag))
					for t := range byTag {
						availableTags = append(availableTags, t)
					}
					fmt.Println(ui.Hint("Available tags: " + strings.Join(availableTags, ", ")))
					return nil
				}
				printEndpoints(endpoints)
			} else {
				printEndpointsByTag(spec)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&tagFilter, "tag", "t", "", "Filter endpoints by tag")

	return cmd
}

// apiExploreCmd provides an interactive API explorer powered by the OpenAPI spec.
func apiExploreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "explore",
		Short: "Interactive API explorer (requires OpenAPI spec)",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}
			if cfg.API == nil {
				return fmt.Errorf("no [api] section in tinker.toml")
			}
			if cfg.API.Spec == "" {
				fmt.Println(ui.Warning("No OpenAPI spec configured"))
				fmt.Println(ui.Hint("Add spec = \"openapi.yaml\" to the [api] section in tinker.toml"))
				fmt.Println(ui.Hint("Then use: tinker api endpoints  to list available endpoints"))
				return nil
			}

			specPath := cfg.API.Spec
			if !filepath.IsAbs(specPath) {
				specPath = filepath.Join(root, specPath)
			}

			spec, err := api.ParseSpec(specPath)
			if err != nil {
				return fmt.Errorf("parsing spec: %w", err)
			}

			opts, err := apiSessionOpts("")
			if err != nil {
				return err
			}

			s, err := api.NewSession(cfg.API, opts...)
			if err != nil {
				return err
			}

			return runAPIExplorer(s, spec)
		},
	}
}

// apiSessionCmd manages HTTP session state (cookies, auth).
func apiSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage HTTP session state (cookies, auth)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "show",
			Short: "Show current session state",
			RunE: func(_ *cobra.Command, _ []string) error {
				_, root, err := loadConfig()
				if err != nil {
					return err
				}

				store, err := api.NewSessionStore(root)
				if err != nil {
					fmt.Println(ui.Dim("  No session data"))
					return nil
				}

				token, authType := store.GetAuth()
				headers := store.GetHeaders()

				if token == "" && len(headers) == 0 {
					fmt.Println(ui.Dim("  No persisted session data"))
					fmt.Println(ui.Hint("Session state is saved automatically when you make API requests"))
					return nil
				}

				fmt.Println()
				fmt.Println("  " + ui.APILabel() + " " + ui.Bold("Session"))
				fmt.Println()
				if token != "" {
					displayToken := token
					if len(token) > 20 {
						displayToken = token[:20] + "..."
					}
					fmt.Println(ui.KeyValue("auth", displayToken))
					fmt.Println(ui.KeyValue("auth_type", authType))
				}
				for k, v := range headers {
					fmt.Println(ui.KeyValue("header:"+k, v))
				}
				fmt.Println()
				return nil
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Clear persisted session state",
			RunE: func(_ *cobra.Command, _ []string) error {
				_, root, err := loadConfig()
				if err != nil {
					return err
				}

				store, err := api.NewSessionStore(root)
				if err != nil {
					fmt.Println(ui.Dim("  No session data to clear"))
					return nil
				}

				if err := store.Clear(); err != nil {
					fmt.Println(ui.Dim("  No session data to clear"))
					return nil
				}

				fmt.Println(ui.Success("Session cleared"))
				return nil
			},
		},
	)

	return cmd
}

func newAPISessionWithOpts(opts ...api.SessionOption) (*api.Session, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return api.NewSession(cfg.API, opts...)
}

// apiSessionOpts builds session options including session persistence.
func apiSessionOpts(jqFilter string) ([]api.SessionOption, error) {
	var opts []api.SessionOption

	if jqFilter != "" {
		opts = append(opts, api.WithJqFilter(jqFilter))
	}

	// Enable session persistence
	_, root, err := loadConfig()
	if err != nil {
		return opts, nil // session persistence is optional
	}
	store, err := api.NewSessionStore(root)
	if err == nil {
		opts = append(opts, api.WithSessionStore(store))
	}

	return opts, nil
}

func parseAPICall(args []string) (method, path, body string) {
	switch len(args) {
	case 1:
		return "GET", args[0], ""
	case 2:
		return args[0], args[1], ""
	default:
		return args[0], args[1], args[2]
	}
}
