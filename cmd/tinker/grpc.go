package main

import (
	"fmt"

	"github.com/mvaliolahi/tinker/internal/grpc"
	"github.com/spf13/cobra"
)

func grpcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grpc",
		Short: "Interact with your project's gRPC services",
	}

	cmd.AddCommand(
		grpcListCmd(),
		grpcDescribeCmd(),
		grpcCallCmd(),
	)

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		s, err := newGRPCSession()
		if err != nil {
			return err
		}
		return s.Interactive()
	}

	return cmd
}

func newGRPCSession() (*grpc.Session, error) {
	cfg, _, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return grpc.NewSession(cfg.GRPC)
}

func grpcListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List gRPC services",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newGRPCSession()
			if err != nil {
				return err
			}
			out, err := s.ListServices()
			fmt.Print(out)
			return err
		},
	}
}

func grpcDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe [service]",
		Short: "Describe a gRPC service",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newGRPCSession()
			if err != nil {
				return err
			}
			out, err := s.Describe(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func grpcCallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "call [method] [data]",
		Short: "Call a gRPC method",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newGRPCSession()
			if err != nil {
				return err
			}
			data := ""
			if len(args) > 1 {
				data = args[1]
			}
			out, err := s.Call(args[0], data)
			fmt.Print(out)
			return err
		},
	}
}
