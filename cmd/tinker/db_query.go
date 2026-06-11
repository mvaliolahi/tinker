package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func dbDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe [table]",
		Short: "Show table schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			out, err := s.Describe(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func dbCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count [table] [where]",
		Short: "Count rows in a table",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			where := ""
			if len(args) > 1 {
				where = args[1]
			}
			out, err := s.Count(args[0], where)
			fmt.Print(out)
			return err
		},
	}
}

func dbFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find [table] [id]",
		Short: "Find a row by ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			out, err := s.Find(args[0], args[1])
			fmt.Print(out)
			return err
		},
	}
}

func dbExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec [sql]",
		Short: "Execute a SQL statement",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			out, err := s.Exec(args[0])
			fmt.Print(out)
			return err
		},
	}
}
