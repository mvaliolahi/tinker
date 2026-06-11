package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mvaliolahi/tinker/internal/config"
	"github.com/mvaliolahi/tinker/internal/runner"
	"github.com/mvaliolahi/tinker/internal/ui"
	"github.com/spf13/cobra"
)

func logCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log [path]",
		Short: "Show log file contents",
		Long:  "Display the contents of a log file. Uses files from [log] config or accepts a manual path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}

			path := pickLogPath(args, cfg)
			if path == "" {
				return fmt.Errorf("no log file specified. Pass a file path or configure [log].files in tinker.toml")
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
		Long:  "Follow a log file in real-time. Uses files from [log] config or accepts a manual path.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, root, err := loadConfig()
			if err != nil {
				return err
			}

			path := pickLogPath(args, cfg)
			if path == "" {
				return fmt.Errorf("no log file specified. Pass a file path or configure [log].files in tinker.toml")
			}

			fullPath := absLogPath(path, root)
			if _, err := os.Stat(fullPath); err != nil {
				return fmt.Errorf("log file not found: %s", fullPath)
			}

			fmt.Println("  " + ui.LogLabel() + " " + ui.Bold("Tailing"))
			fmt.Println(ui.KeyValue("file", fullPath))
			fmt.Println(ui.Dim("  Ctrl+C to stop"))
			fmt.Println()

			// Use system `tail -f` if available for proper follow behavior
			if _, err := exec.LookPath("tail"); err == nil {
				return runner.Interactive("tail", "-f", fullPath)
			}

			// Fallback: manual follow implementation
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
				fmt.Println(ui.Hint("Add [log] section to tinker.toml:"))
				fmt.Println(ui.Dim("  [log]"))
				fmt.Println(ui.Dim("  " + `files = ["app.log", "logs/error.log"]`))
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

// pickLogPath determines which log file to use.
// Priority: 1) manual path arg, 2) first file from config.
func pickLogPath(args []string, cfg *config.Config) string {
	if len(args) > 0 {
		return args[0]
	}
	if cfg.Log != nil && len(cfg.Log.Files) > 0 {
		return cfg.Log.Files[0]
	}
	return ""
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

	// Seek to end
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

// colorizeLogLine applies color to common log patterns.
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
