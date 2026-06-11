package main

import (
	"fmt"

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

	cmd.Flags().StringVarP(&jqFilter, "jq", "q", "", "Filter response with jq expression")

	cmd.AddCommand(&cobra.Command{
		Use:   "interactive",
		Short: "Open interactive HTTP session",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newAPISession()
			if err != nil {
				return err
			}
			fmt.Println("  " + ui.APILabel() + " " + ui.Bold("Interactive session"))
			fmt.Println(ui.KeyValue("base", s.BaseURL))
			fmt.Println()
			return s.Interactive()
		},
	})

	cmd.RunE = func(_ *cobra.Command, args []string) error {
		var opts []api.SessionOption
		if jqFilter != "" {
			opts = append(opts, api.WithJqFilter(jqFilter))
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

func newAPISession() (*api.Session, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return api.NewSession(cfg.API)
}

func newAPISessionWithOpts(opts ...api.SessionOption) (*api.Session, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return api.NewSession(cfg.API, opts...)
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
