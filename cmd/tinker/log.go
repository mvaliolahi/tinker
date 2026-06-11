package main

import (
        "bufio"
        "encoding/json"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "sort"
        "strconv"
        "strings"
        "time"

        "github.com/mvaliolahi/tinker/internal/config"
        "github.com/mvaliolahi/tinker/internal/contract"
        "github.com/mvaliolahi/tinker/internal/detect"
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
                                fmt.Println(formatLogLine(scanner.Text()))
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

                        // Use manual tail to format/colorize each line
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
                fmt.Println(formatLogLine(strings.TrimRight(line, "\n")))
        }
}

// formatLogLine detects JSON log lines and formats them, falls back to text colorization.
func formatLogLine(line string) string {
        line = strings.TrimSpace(line)
        if line == "" {
                return ""
        }

        // Try JSON first
        if line[0] == '{' {
                var obj map[string]interface{}
                if json.Unmarshal([]byte(line), &obj) == nil {
                        return formatJSONLog(obj)
                }
        }

        // Fallback: text-based colorization
        return colorizeTextLog(line)
}

// formatJSONLog renders a structured JSON log line as a clean, colorized single line.
//
// Input:  {"ts":"2026-06-11T20:52:08.522Z","level":"warn","msg":"http request","method":"POST","path":"/api/v1/auth/register","status":409,"duration":734680}
// Output: 2026-06-11 20:52:08 WARN  POST /api/v1/auth/register → 409  734.7ms
func formatJSONLog(obj map[string]interface{}) string {
        var sb strings.Builder

        // Timestamp
        ts := fieldStr(obj, "ts", "time", "t", "timestamp", "datetime")
        if ts != "" {
                sb.WriteString(ui.Dim(formatTimestamp(ts)))
                sb.WriteString(" ")
        }

        // Level
        level := fieldStr(obj, "level", "lvl", "severity")
        if level != "" {
                sb.WriteString(formatLevel(level))
                sb.WriteString(" ")
        }

        // Check if this is an HTTP request log
        method := fieldStr(obj, "method")
        path := fieldStr(obj, "path", "url", "uri")
        status := fieldInt(obj, "status", "status_code", "statusCode")
        duration := fieldInt(obj, "duration", "dur", "latency")

        if method != "" || path != "" {
                // HTTP request log format: METHOD /path → status duration
                if method != "" {
                        sb.WriteString(ui.HTTPMethod(method))
                        sb.WriteString(" ")
                }
                if path != "" {
                        sb.WriteString(ui.Accent(path))
                }
                if status > 0 {
                        sb.WriteString(" ")
                        sb.WriteString(ui.Dim("→"))
                        sb.WriteString(" ")
                        sb.WriteString(formatStatus(status))
                }
                if duration > 0 {
                        sb.WriteString("  ")
                        sb.WriteString(ui.Dim(formatDuration(duration)))
                }
        } else {
                // General log: message
                msg := fieldStr(obj, "msg", "message", "text")
                if msg != "" {
                        sb.WriteString(msg)
                }
        }

        // Event
        event := fieldStr(obj, "event")
        if event != "" && method == "" {
                sb.WriteString(" ")
                sb.WriteString(ui.Dim(event))
        }

        // Signal (for shutdown logs)
        signal := fieldStr(obj, "signal")
        if signal != "" {
                sb.WriteString(" ")
                sb.WriteString(ui.Warning(signal))
        }

        // Remaining fields as key=value (skip already-shown fields)
        shown := map[string]bool{
                "ts": true, "time": true, "t": true, "timestamp": true, "datetime": true,
                "level": true, "lvl": true, "severity": true,
                "msg": true, "message": true, "text": true,
                "method": true, "path": true, "url": true, "uri": true,
                "status": true, "status_code": true, "statusCode": true,
                "duration": true, "dur": true, "latency": true,
                "event": true, "signal": true,
        }

        keys := make([]string, 0, len(obj))
        for k := range obj {
                if !shown[k] {
                        keys = append(keys, k)
                }
        }
        sort.Strings(keys)

        if len(keys) > 0 {
                sb.WriteString("  ")
                for i, k := range keys {
                        if i > 0 {
                                sb.WriteString("  ")
                        }
                        v := obj[k]
                        sb.WriteString(ui.Dim(k + "="))
                        sb.WriteString(formatValue(v))
                }
        }

        return sb.String()
}

// formatLevel returns a colorized, left-aligned level string.
func formatLevel(level string) string {
        upper := strings.ToUpper(level)
        padded := fmt.Sprintf("%-5s", upper)

        switch upper {
        case "ERROR", "FATAL", "PANIC", "CRITICAL":
                return ui.Error(padded)
        case "WARN", "WARNING":
                return ui.Warning(padded)
        case "INFO", "NOTICE":
                return ui.Success(padded)
        case "DEBUG", "TRACE":
                return ui.Dim(padded)
        default:
                return padded
        }
}

// formatStatus returns a colorized HTTP status code.
func formatStatus(code int64) string {
        s := fmt.Sprintf("%d", code)
        switch {
        case code >= 200 && code < 300:
                return ui.Success(s)
        case code >= 300 && code < 400:
                return ui.Accent(s)
        case code >= 400 && code < 500:
                return ui.Warning(s)
        default:
                return ui.Error(s)
        }
}

// formatDuration formats nanoseconds as a human-readable duration.
func formatDuration(ns int64) string {
        switch {
        case ns >= 1_000_000_000:
                return fmt.Sprintf("%.1fs", float64(ns)/1_000_000_000)
        case ns >= 1_000_000:
                return fmt.Sprintf("%.1fms", float64(ns)/1_000_000)
        case ns >= 1_000:
                return fmt.Sprintf("%.1fµs", float64(ns)/1_000)
        default:
                return fmt.Sprintf("%dns", ns)
        }
}

// formatTimestamp converts an ISO timestamp to a friendlier format.
func formatTimestamp(ts string) string {
        for _, layout := range []string{
                time.RFC3339Nano,
                time.RFC3339,
                "2006-01-02T15:04:05Z07:00",
                "2006-01-02T15:04:05",
                "2006-01-02 15:04:05",
        } {
                if t, err := time.Parse(layout, ts); err == nil {
                        return t.Format("2006-01-02 15:04:05")
                }
        }
        return ts
}

// formatValue renders a JSON value as a short string for key=value display.
func formatValue(v interface{}) string {
        switch val := v.(type) {
        case string:
                if val == "" {
                        return ui.Dim(`""`)
                }
                if len(val) > 60 {
                        return ui.Dim(val[:57] + "...")
                }
                return val
        case float64:
                if val == float64(int64(val)) {
                        return fmt.Sprintf("%d", int64(val))
                }
                return fmt.Sprintf("%.2f", val)
        case bool:
                if val {
                        return ui.Success("true")
                }
                return ui.Error("false")
        case nil:
                return ui.Dim("null")
        default:
                s := fmt.Sprintf("%v", val)
                if len(s) > 60 {
                        s = s[:57] + "..."
                }
                return ui.Dim(s)
        }
}

// fieldStr extracts a string value from a map, trying multiple key names.
func fieldStr(obj map[string]interface{}, keys ...string) string {
        for _, k := range keys {
                if v, ok := obj[k]; ok {
                        if s, ok := v.(string); ok {
                                return s
                        }
                }
        }
        return ""
}

// fieldInt extracts an int64 value from a map, trying multiple key names.
func fieldInt(obj map[string]interface{}, keys ...string) int64 {
        for _, k := range keys {
                if v, ok := obj[k]; ok {
                        switch n := v.(type) {
                        case float64:
                                return int64(n)
                        case int64:
                                return n
                        case int:
                                return int64(n)
                        }
                }
        }
        return 0
}

// colorizeTextLog handles non-JSON log lines (plain text).
func colorizeTextLog(line string) string {
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
