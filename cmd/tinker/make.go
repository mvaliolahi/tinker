package main

import (
	"fmt"

	tmaker "github.com/mvaliolahi/tinker/internal/make"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func makeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make [target] [args...]",
		Short: "Run Makefile targets",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			dir, err := resolveDir()
			if err != nil {
				return err
			}
			if !tmaker.HasMakefile(dir) {
				return fmt.Errorf("no Makefile found in %s", dir)
			}
			fmt.Println("  " + ui.MakeLabel() + " " + ui.Bold("make "+args[0]))
			fmt.Println()
			return tmaker.Run(dir, args[0], args[1:])
		},
	}

	cmd.AddCommand(makeListCmd())

	return cmd
}

func makeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Makefile targets",
		RunE: func(_ *cobra.Command, _ []string) error {
			dir, err := resolveDir()
			if err != nil {
				return err
			}
			if !tmaker.HasMakefile(dir) {
				return fmt.Errorf("no Makefile found in %s", dir)
			}
			fmt.Println()
			fmt.Println("  " + ui.MakeLabel() + " " + ui.Bold("Targets"))
			fmt.Println()
			fmt.Print(ui.NumberedList(tmaker.Targets(dir)))
			return nil
		},
	}
}
