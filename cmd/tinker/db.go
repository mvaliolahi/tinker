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
		dbCountCmd(),
		dbFindCmd(),
		dbExecCmd(),
	)

	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		s, err := newDBSession()
		if err != nil {
			return err
		}
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
	fmt.Println()
	fmt.Println(ui.Dim("  \\d [table]  describe · \\dt  list tables · \\q  quit · Ctrl+C  cancel"))
	fmt.Println(ui.Dim("  Note: use \\d not .schema, \\dt not .tables (usql, not sqlite3)"))
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
			printDBInfo(s)
			return s.Connect()
		},
	}
}

func dbTablesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tables",
		Short: "List all tables",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := newDBSession()
			if err != nil {
				return err
			}
			fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Tables"))
			fmt.Println()
			out, err := s.Tables()
			fmt.Print(out)
			return err
		},
	}
}
