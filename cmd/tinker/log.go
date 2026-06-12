package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/contract"
	"github.com/mvaliolahi/tinker/internal/detect"
	"github.com/mvaliolahi/tinker/internal/logfmt"
	"github.com/mvaliolahi/tinker/internal/logtail"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
	var tailN int
	var levelFilter string
	var grepFilter string

	cmd := &cobra.Command{
		Use:   "log [path]",
		Short: "Show log file contents",
		Long:  "Display the contents of a log file. Uses files from [log] config, scans for logs, or accepts a manual path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}

			path, err := resolveLogPath(args, cfg, root)
			if err != nil {
				return err
			}

			fullPath := absLogPath(path, root)
			if _, err := os.Stat(fullPath); err != nil {
				return fmt.Errorf("log file not found: %s", fullPath)
			}

			fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Log"))
			fmt.Println(ui.KeyValue("file", fullPath))
			printFilterInfo(levelFilter, grepFilter)
			fmt.Println()

			return streamWithFilter(fullPath, tailN, levelFilter, grepFilter)
		},
	}

	cmd.Flags().IntVarP(&tailN, "tail", "n", 0, "Show last N lines (0 = all)")
	cmd.Flags().StringVarP(&levelFilter, "level", "l", "", "Filter by log level (error, warn, info, debug)")
	cmd.Flags().StringVarP(&grepFilter, "grep", "g", "", "Filter by text pattern")

	cmd.AddCommand(&cobra.Command{
		Use:   "tail [path]",
		Short: "Tail (follow) a log file in real-time",
		Long:  "Follow a log file in real-time. Uses files from [log] config, scans for logs, or accepts a manual path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}

			path, err := resolveLogPath(args, cfg, root)
			if err != nil {
				return err
			}

			fullPath := absLogPath(path, root)
			if _, err := os.Stat(fullPath); err != nil {
				return fmt.Errorf("log file not found: %s", fullPath)
			}

			fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Tailing"))
			fmt.Println(ui.KeyValue("file", fullPath))
			fmt.Println(ui.Dim("  Ctrl+C to stop"))
			fmt.Println()

			return logtail.Follow(fullPath, logtail.DefaultTailLines)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List configured log files",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}

			fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Log Files"))
			fmt.Println()

			if cfg.Log == nil || len(cfg.Log.Files) == 0 {
				fmt.Println(ui.Dim("  No log files configured."))
				fmt.Println()
				fmt.Println(ui.Hint("Run 'tinker log' to scan and select log files"))
				return nil
			}

			for i, f := range cfg.Log.Files {
				fullPath := absLogPath(f, root)
				exists := ui.Success("found")
				if _, err := os.Stat(fullPath); err != nil {
					exists = ui.Warning("missing")
				}
				fmt.Printf("  %s %s  %s\n", ui.Dim(fmt.Sprintf("%d.", i+1)), ui.Bold(f), exists)
			}

			fmt.Println()
			fmt.Println(ui.Hint("tinker log          Show first log file"))
			fmt.Println(ui.Hint("tinker log tail     Follow log in real-time"))
			fmt.Println(ui.Hint("tinker log -n 50    Show last 50 lines"))
			fmt.Println(ui.Hint("tinker log -l error Filter by level"))
			fmt.Println(ui.Hint("tinker log -g timeout  Filter by text"))
			return nil
		},
	})

	return cmd
}

func printFilterInfo(level, grep string) {
	if level != "" {
		fmt.Println(ui.KeyValue("level", level))
	}
	if grep != "" {
		fmt.Println(ui.KeyValue("grep", grep))
	}
}

func streamWithFilter(path string, tailN int, level, grep string) error {
	if tailN > 0 {
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening log file: %w", err)
		}
		defer f.Close()

		lines := logtail.ReadLastLines(f, tailN)
		for _, line := range lines {
			if matchesFilter(line, level, grep) {
				fmt.Println(logfmt.FormatLine(line))
			}
		}
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if matchesFilter(line, level, grep) {
			fmt.Println(logfmt.FormatLine(line))
		}
	}
	return scanner.Err()
}

func matchesFilter(line, level, grep string) bool {
	if level != "" && !logfmt.MatchLevel(line, level) {
		return false
	}
	if grep != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(grep)) {
		return false
	}
	return true
}

func resolveLogPath(args []string, cfg *config.Config, root string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if cfg.Log != nil && len(cfg.Log.Files) > 0 {
		return cfg.Log.Files[0], nil
	}
	return scanAndSelect(root)
}

func scanAndSelect(root string) (string, error) {
	fmt.Println()
	fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("No log files configured"))
	fmt.Println(ui.Dim("  Scanning project for log files..."))
	fmt.Println()

	found := detect.ScanLogFiles(root)
	if len(found) == 0 {
		fmt.Println(ui.Warning("No log files found in this project."))
		fmt.Println()
		fmt.Println(ui.Hint("Pass a file path directly:"))
		fmt.Println(ui.Dim("  tinker log /path/to/file.log"))
		return "", fmt.Errorf("no log files found")
	}

	fmt.Println(ui.Bold("  Found log files:"))
	fmt.Println()
	for i, f := range found {
		fullPath := absLogPath(f, root)
		size := fileSize(fullPath)
		fmt.Printf("  %s %s  %s\n",
			ui.Dim(fmt.Sprintf("%d.", i+1)),
			ui.Bold(f),
			ui.Dim(size),
		)
	}
	fmt.Println()

	fmt.Print(ui.Accent("  Select") + " " + ui.Dim("(number or path, Enter=all): "))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var selected []string

	if input == "" {
		selected = found
	} else if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= len(found) {
		selected = []string{found[num-1]}
	} else {
		if _, err := os.Stat(input); err != nil {
			return "", fmt.Errorf("file not found: %s", input)
		}
		selected = []string{input}
	}

	if err := contract.SaveLogConfig(selected, root); err != nil {
		fmt.Println(ui.Warning("Could not save to tinker.toml: " + err.Error()))
	} else {
		fmt.Println()
		fmt.Println(ui.Success("Saved to tinker.toml"))
		for _, f := range selected {
			fmt.Println(ui.Bullet("log", f))
		}
	}

	fmt.Println()
	return selected[0], nil
}

func fileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	size := info.Size()
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func absLogPath(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}
