package main

import (
        "fmt"
        "os"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/mvaliolahi/tinker/internal/db"
        "github.com/mvaliolahi/tinker/internal/ui"
        "github.com/spf13/cobra"
)

func dbSeedCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "seed [path]",
                Short: "Run seed files against the database",
                Long:  "Execute SQL seed files against the database. Pass a directory to run all .sql files, or a single file path.",
                Args:  cobra.MaximumNArgs(1),
                RunE: func(_ *cobra.Command, args []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()

                        path := "seed"
                        if len(args) > 0 {
                                path = args[0]
                        }

                        // Resolve relative to project root
                        _, root, err := loadConfig()
                        if err != nil {
                                return err
                        }
                        if !isAbs(path) {
                                path = root + "/" + path
                        }

                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Seed"))
                        fmt.Println(ui.Dim("  path: "+path))
                        fmt.Println()

                        if err := s.Seed(path); err != nil {
                                fmt.Println(ui.Error(err.Error()))
                                os.Exit(1)
                        }

                        fmt.Println(ui.Success("Seed completed"))
                        return nil
                },
        }
}

func dbMigrateCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "migrate",
                Short: "Run database migrations",
                Long:  "Run SQL migration files with version tracking. Supports up, down, and status subcommands.",
        }

        cmd.AddCommand(
                dbMigrateUpCmd(),
                dbMigrateDownCmd(),
                dbMigrateStatusCmd(),
        )

        return cmd
}

func dbMigrateUpCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "up",
                Short: "Run pending migrations",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()

                        dir, err := migrateDir()
                        if err != nil {
                                return err
                        }

                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Migrate Up"))
                        fmt.Println(ui.Dim("  dir: "+dir))
                        fmt.Println()

                        applied, err := s.MigrateUp(dir)
                        if err != nil {
                                fmt.Println(ui.Error(err.Error()))
                                os.Exit(1)
                        }

                        if len(applied) == 0 {
                                fmt.Println(ui.Dim("  No pending migrations"))
                        } else {
                                for _, name := range applied {
                                        fmt.Println("  " + ui.Success(name))
                                }
                                fmt.Println()
                                fmt.Printf("  Applied %d migration(s)\n", len(applied))
                        }
                        return nil
                },
        }
}

func dbMigrateDownCmd() *cobra.Command {
        return &cobra.Command{
                Use:     "down",
                Aliases: []string{"rollback"},
                Short:   "Rollback the last migration",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()

                        dir, err := migrateDir()
                        if err != nil {
                                return err
                        }

                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Migrate Down"))
                        fmt.Println()

                        name, err := s.MigrateDown(dir)
                        if err != nil {
                                fmt.Println(ui.Error(err.Error()))
                                os.Exit(1)
                        }

                        if name == "" {
                                fmt.Println(ui.Dim("  No migrations to rollback"))
                        } else {
                                fmt.Println("  " + ui.Success("Rolled back: "+name))
                        }
                        return nil
                },
        }
}

func dbMigrateStatusCmd() *cobra.Command {
        return &cobra.Command{
                Use:     "status",
                Aliases: []string{"st"},
                Short:   "Show migration status",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()

                        dir, err := migrateDir()
                        if err != nil {
                                return err
                        }

                        fmt.Println("  " + ui.DBLabel() + " " + ui.Bold("Migration Status"))
                        fmt.Println()

                        status, err := s.MigrateStatus(dir)
                        if err != nil {
                                fmt.Println(ui.Error(err.Error()))
                                os.Exit(1)
                        }

                        if len(status) == 0 {
                                fmt.Println(ui.Dim("  No migrations found"))
                                return nil
                        }

                        headers := []string{"Version", "Name", "Status"}
                        var rows [][]string
                        for _, m := range status {
                                st := ui.Success("applied")
                                if !m.Applied {
                                        st = ui.Warning("pending")
                                }
                                rows = append(rows, []string{m.Version, m.Name, st})
                        }
                        fmt.Print(renderCLIResult(headers, rows))
                        return nil
                },
        }
}

func dbExploreCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "explore",
                Short: "Interactive database browser (TUI)",
                RunE: func(_ *cobra.Command, _ []string) error {
                        s, err := newDBSession()
                        if err != nil {
                                return err
                        }
                        defer s.Close()

                        model := db.NewExploreModel(s)
                        p := tea.NewProgram(model, tea.WithAltScreen())
                        if _, err := p.Run(); err != nil {
                                return fmt.Errorf("TUI error: %w", err)
                        }
                        return nil
                },
        }
}

func migrateDir() (string, error) {
        _, root, err := loadConfig()
        if err != nil {
                return "", err
        }

        // Check common migration directories
        candidates := []string{
                root + "/migrations",
                root + "/migrate",
                root + "/db/migrations",
                root + "/sql/migrations",
                root + "/backend/migrations",
                root + "/backend/migrate",
        }

        for _, dir := range candidates {
                if info, err := os.Stat(dir); err == nil && info.IsDir() {
                        return dir, nil
                }
        }

        // Default to migrations/
        return root + "/migrations", nil
}

func isAbs(path string) bool {
        return len(path) > 0 && path[0] == '/'
}

// renderCLIResult renders a simple table for CLI output.
func renderCLIResult(headers []string, rows [][]string) string {
        return ui.Table(headers, rows)
}
