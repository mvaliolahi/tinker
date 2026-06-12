package main

import (
        "fmt"

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
                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Count "+args[0]))
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
                        out, err := s.Exec(args[0])
                        fmt.Print(out)
                        return err
                },
        }
}
