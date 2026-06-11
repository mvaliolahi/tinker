package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/contract"
	"github.com/mvaliolahi/tinker/internal/detect"
	"github.com/mvaliolahi/tinker/internal/runner"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
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
			fmt.Println()

			f, err := os.Open(fullPath)
			if err != nil {
				return fmt.Errorf("opening log file: %w", err)
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for scanner.Scan() {
				fmt.Println(colorizeLogLine(scanner.Text()))
			}
			return scanner.Err()
		},
	}

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

			if _, err := exec.LookPath("tail"); err == nil {
				return runner.Interactive("tail", "-f", fullPath)
			}

			return manualTail(fullPath)
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
			fmt.Println(ui.Hint("tinker log path.log Show specific file"))
			return nil
		},
	})

	return cmd
}

// resolveLogPath determines which log file to use.
// Priority: 1) manual path arg, 2) first file from config, 3) scan + interactive selection.
func resolveLogPath(args []string, cfg *config.Config, root string) (string, error) {
	// 1. Manual path argument
	if len(args) > 0 {
		return args[0], nil
	}

	// 2. Configured log files
	if cfg.Log != nil && len(cfg.Log.Files) > 0 {
		return cfg.Log.Files[0], nil
	}

	// 3. Scan project for log files
	return scanAndSelect(root)
}

// scanAndSelect scans the project for *.log files, presents them to the user,
// saves the selection to tinker.toml, and returns the chosen path.
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

	// Show found files with file sizes
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

	// Prompt for selection
	fmt.Print(ui.Accent("  Select") + " " + ui.Dim("(number or path, Enter=all): "))

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var selected []string

	if input == "" {
		// Select all
		selected = found
	} else if num, err := strconv.Atoi(input); err == nil && num >= 1 && num <= len(found) {
		// Select by number
		selected = []string{found[num-1]}
	} else {
		// Treat as a file path
		if _, err := os.Stat(input); err != nil {
			return "", fmt.Errorf("file not found: %s", input)
		}
		selected = []string{input}
	}

	// Save to tinker.toml
	if err := contract.SaveLogConfig(selected, root); err != nil {
		// Non-fatal — still proceed to show the log
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

// absLogPath resolves a potentially relative path against the project root.
func absLogPath(path, root string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, path)
}

func manualTail(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(250 * time.Millisecond)
				continue
			}
			return err
		}
		fmt.Print(colorizeLogLine(strings.TrimRight(line, "\n")))
		fmt.Println()
	}
}

func colorizeLogLine(line string) string {
	line = colorizeLevels(line)
	line = colorizeTimestamps(line)
	return line
}

func colorizeLevels(line string) string {
	replacements := []struct {
		pattern string
		style   func(string) string
	}{
		{"[ERROR]", ui.Error}, {"[FATAL]", ui.Error}, {"[PANIC]", ui.Error}, {"[CRITICAL]", ui.Error},
		{"[WARN]", ui.Warning}, {"[WARNING]", ui.Warning},
		{"[INFO]", ui.Success}, {"[NOTICE]", ui.Success},
		{"[DEBUG]", ui.Dim}, {"[TRACE]", ui.Dim},
		{"ERROR:", ui.Error}, {"FATAL:", ui.Error}, {"PANIC:", ui.Error},
		{"WARN:", ui.Warning}, {"WARNING:", ui.Warning},
		{"INFO:", ui.Success}, {"NOTICE:", ui.Success},
		{"DEBUG:", ui.Dim}, {"TRACE:", ui.Dim},
	}

	for _, r := range replacements {
		if strings.Contains(line, r.pattern) {
			line = strings.ReplaceAll(line, r.pattern, r.style(r.pattern))
		}
	}
	return line
}

func colorizeTimestamps(line string) string {
	if len(line) > 4 {
		if line[:4] >= "2000" && line[:4] <= "2099" && len(line) > 4 && (line[4] == '-' || line[4] == '/') {
			end := 25
			if len(line) < end {
				end = len(line)
			}
			ts := line[:end]
			rest := ""
			if len(line) > end {
				rest = line[end:]
			}
			return ui.Dim(ts) + rest
		}
	}
	return line
}
