package main

import (
	"fmt"
	"time"

	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func dbDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "describe [table]",
		Aliases: []string{"desc"},
		Short:   "Show table schema",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Describe "+args[0]))
			fmt.Println()
			out, err := s.Describe(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func dbIndexesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "indexes [table]",
		Aliases: []string{"idx"},
		Short:   "Show table indexes",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Indexes "+args[0]))
			fmt.Println()
			out, err := s.Indexes(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func dbSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "schema [table]",
		Aliases: []string{"s"},
		Short:   "Show CREATE TABLE statement",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Schema "+args[0]))
			fmt.Println()
			out, err := s.Schema(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func dbCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "count [table] [where]",
		Aliases: []string{"c"},
		Short:   "Count rows in a table",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			where := ""
			if len(args) > 1 {
				where = args[1]
			}
			label := "Count " + args[0]
			if where != "" {
				label += " WHERE " + where
			}
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold(label))
			fmt.Println()
			out, err := s.Count(args[0], where)
			fmt.Print(out)
			return err
		},
	}
}

func dbFindCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "find [table] [id]",
		Aliases: []string{"f"},
		Short:   "Find a row by ID",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Find "+args[0]+" #"+args[1]))
			fmt.Println()
			out, err := s.Find(args[0], args[1])
			fmt.Print(out)
			return err
		},
	}
}

func dbExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "exec [sql]",
		Aliases: []string{"e", "sql"},
		Short:   "Execute a SQL statement",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Exec"))
			fmt.Println(ui.Dim("  " + args[0]))
			fmt.Println()
			out, err := s.ExecFormatted(args[0])
			fmt.Print(out)
			return err
		},
	}
}

func dbPingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Test database connectivity",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			start := time.Now()
			if err := s.Ping(); err != nil {
				fmt.Println("  " + ui.DBLabel() + " " + ui.Error("unreachable"))
				fmt.Println(ui.Dim("  "+err.Error()))
				return nil
			}
			elapsed := time.Since(start)
			fmt.Println("  " + ui.DBLabel() + " " + ui.Success("reachable") + " " + ui.Dim(elapsed.Round(time.Millisecond).String()))
			return nil
		},
	}
}

func dbSizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "size",
		Short: "Show table row counts",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			defer s.Close()
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Size"))
			fmt.Println()
			out, err := s.Size()
			fmt.Print(out)
			return err
		},
	}
}
