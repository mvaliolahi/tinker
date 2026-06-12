package main

import (
        "fmt"

        "github.com/mvaliolahi/tinker/internal/db"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

func dbCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "db",
                Short: "Interact with your project's database",
        }

        cmd.AddCommand(
                dbConnectCmd(),
                dbTablesCmd(),
                dbDescribeCmd(),
                dbIndexesCmd(),
                dbSchemaCmd(),
                dbCountCmd(),
                dbFindCmd(),
                dbExecCmd(),
                dbPingCmd(),
                dbSizeCmd(),
                dbSeedCmd(),
                dbMigrateCmd(),
                dbExploreCmd(),
        )

        cmd.RunE = func(_ *cobra.Command, _ []string) error {
                s, err := newDBSession()
                if err != nil {
                        return err
                }
                defer s.Close()
                printDBInfo(s)
                return s.Connect()
        }

        return cmd
}

func newDBSession() (*db.Session, error) {
        cfg, _, err := loadConfig()
        if err != nil {
                return nil, err
        }
        return db.NewSession(cfg.Database)
}

func printDBInfo(s *db.Session) {
        fmt.Println()
        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Database"))
        fmt.Println(ui.KeyValue("type", s.Type))
        fmt.Println(ui.KeyValue("driver", s.Driver))
        fmt.Println(ui.KeyValue("cli", s.CLI()))
        if s.HasNativeConn() {
                fmt.Println(ui.KeyValue("native", "connected ✓"))
        } else {
                fmt.Println(ui.KeyValue("native", "not available (CLI only)"))
        }
        fmt.Println()

        cli := s.CLI()
        switch cli {
        case "litecli", "pgcli", "mycli":
                fmt.Println(ui.Dim("  Syntax highlighting + autocomplete enabled"))
                fmt.Println(ui.Dim("  Type \\q or Ctrl+D to quit"))
        default:
                if modern := s.ModernCLIAvailable(); modern != "" {
                        fmt.Println(ui.Hint("Install " + modern + " for syntax highlighting + autocomplete:"))
                        fmt.Println(ui.Dim("  pip3 install --user " + modern))
                }
                fmt.Println(ui.Dim("  Type \\q or Ctrl+D to quit · Ctrl+C to cancel input"))
        }
        fmt.Println()
}

func dbConnectCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "connect",
                Short: "Open interactive database session",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()
                        printDBInfo(s)
                        return s.Connect()
                },
        }
}

func dbTablesCmd() *cobra.Command {
        return &cobra.Command{
                Use:     "tables",
                Aliases: []string{"ls"},
                Short:   "List all tables",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()
                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Tables"))
                        fmt.Println()
                        out, err := s.Tables()
                        fmt.Print(out)
                        return err
                },
        }
}
