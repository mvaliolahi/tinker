package main

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/api"
	"github.com/spf13/cobra"
)

func apiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api [method] [path] [body]",
		Short: "Call your project's HTTP API",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "interactive",
		Short: "Open interactive HTTP session",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newAPISession()
			if err != nil {
				return err
			}
			return s.Interactive()
		},
	})

	cmd.RunE = func(_ *cobra.Command, args []string) error {
		s, err := newAPISession()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return s.Interactive()
		}

		method, path, body := parseAPICall(args)
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

func parseAPICall(args []string) (method, path, body string) {
	switch len(args) {
	case 1:
		return "GET", args[0], ""
	case 2:
		return args[0], args[1], ""
	case 3:
		return args[0], args[1], args[2]
	default:
		return args[0], args[1], args[2]
	}
}
