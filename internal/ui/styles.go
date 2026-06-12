package ui

import (
        "fmt"
        "strings"

        "github.com/charmbracelet/lipgloss"
)

var (
        bold    = lipgloss.NewStyle().Bold(true)
        faint   = lipgloss.NewStyle().Faint(true)
        accent  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
        success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
        error_  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
        warning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
        info    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))

        labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
        valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
        dimStyle   = lipgloss.NewStyle().Faint(true)

        badge = lipgloss.NewStyle().Bold(true).Padding(0, 1)

        dbBadge   = badge.Foreground(lipgloss.Color("6")).Background(lipgloss.Color("23"))
        apiBadge  = badge.Foreground(lipgloss.Color("3")).Background(lipgloss.Color("58"))
        grpcBadge = badge.Foreground(lipgloss.Color("5")).Background(lipgloss.Color("53"))
        makeBadge = badge.Foreground(lipgloss.Color("2")).Background(lipgloss.Color("22"))
        runBadge  = badge.Foreground(lipgloss.Color("12")).Background(lipgloss.Color("17"))
        logBadge  = badge.Foreground(lipgloss.Color("11")).Background(lipgloss.Color("58"))
        cfgBadge  = badge.Foreground(lipgloss.Color("14")).Background(lipgloss.Color("17"))

        panelBorder = lipgloss.NewStyle().
                        Border(lipgloss.RoundedBorder()).
                        BorderForeground(lipgloss.Color("8")).
                        Padding(0, 1)

        accentBorder = lipgloss.NewStyle().
                        Border(lipgloss.RoundedBorder()).
                        BorderForeground(lipgloss.Color("12")).
                        Padding(0, 1)
)

func Success(text string) string {
        return success.Bold(true).Render("✓") + " " + success.Render(text)
}
func Error(text string) string {
        return error_.Bold(true).Render("✗") + " " + error_.Render(text)
}
func Warning(text string) string {
        return warning.Bold(true).Render("⚠") + " " + warning.Render(text)
}
func Dim(text string) string    { return dimStyle.Render(text) }
func Bold(text string) string   { return bold.Render(text) }
func Accent(text string) string { return accent.Render(text) }

func DBLabel() string   { return dbBadge.Render(" DB ") }
func APILabel() string  { return apiBadge.Render(" API ") }
func GRPCLabel() string { return grpcBadge.Render(" GRPC ") }
func MakeLabel() string { return makeBadge.Render(" MAKE ") }
func RunLabel() string  { return runBadge.Render(" RUN ") }
func LogLabel() string  { return logBadge.Render(" LOG ") }
func CfgLabel() string  { return cfgBadge.Render(" CFG ") }

func KeyValue(label, value string) string {
        return "  " + labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func Bullet(label, value string) string {
        return "  " + accent.Render("▪") + " " + labelStyle.Render(label) + " " + valueStyle.Render(value)
}

func Section(title string) string {
        line := strings.Repeat("─", 40)
        return "\n" + accent.Render(line) + "\n  " + info.Bold(true).Render(title) + "\n" + accent.Render(line)
}

func Panel(title, content string) string {
        header := info.Bold(true).Render(" " + title + " ")
        return panelBorder.Render(header + "\n" + content)
}

func AccentPanel(title, content string) string {
        header := accent.Bold(true).Render(" " + title + " ")
        return accentBorder.Render(header + "\n" + content)
}

func Banner(version string) string {
        logo := fmt.Sprintf(
                "  %s\n  %s\n  %s",
                accent.Bold(true).Render("╔═══════════════════════╗"),
                accent.Bold(true).Render("║")+"  "+bold.Render("⚡ Tinker")+" "+faint.Render("v"+version)+"  "+accent.Bold(true).Render("║"),
                accent.Bold(true).Render("╚═══════════════════════╝"),
        )
        return logo
}

func DepStatus(name string, installed bool, purpose string) string {
        icon := success.Bold(true).Render("✓")
        status := success.Render("installed")
        if !installed {
                icon = error_.Bold(true).Render("✗")
                status = error_.Render("missing")
        }
        return fmt.Sprintf("  %s %-10s %-12s %s", icon, valueStyle.Render(name), status, labelStyle.Render(purpose))
}

func HTTPMethod(method string) string {
        var s lipgloss.Style
        switch strings.ToUpper(method) {
        case "GET":
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
        case "POST":
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
        case "PUT":
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
        case "PATCH":
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
        case "DELETE":
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
        default:
                s = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
        }
        return s.Render(strings.ToUpper(method))
}

func Table(headers []string, rows [][]string) string {
        if len(rows) == 0 {
                return ""
        }

        colWidths := make([]int, len(headers))
        for i, h := range headers {
                colWidths[i] = len(h)
        }
        for _, row := range rows {
                for i, cell := range row {
                        if i < len(colWidths) && len(cell) > colWidths[i] {
                                colWidths[i] = len(cell)
                        }
                }
        }

        var sb strings.Builder
        for i, h := range headers {
                sb.WriteString(labelStyle.Render(padRight(h, colWidths[i])))
                if i < len(headers)-1 {
                        sb.WriteString("  ")
                }
        }
        sb.WriteString("\n")

        for i := range headers {
                sb.WriteString(dimStyle.Render(strings.Repeat("─", colWidths[i])))
                if i < len(headers)-1 {
                        sb.WriteString("  ")
                }
        }
        sb.WriteString("\n")

        for _, row := range rows {
                for i, cell := range row {
                        if i < len(colWidths) {
                                sb.WriteString(valueStyle.Render(padRight(cell, colWidths[i])))
                        }
                        if i < len(headers)-1 {
                                sb.WriteString("  ")
                        }
                }
                sb.WriteString("\n")
        }

        return sb.String()
}

func NumberedList(items []string) string {
        var sb strings.Builder
        for i, item := range items {
                num := accent.Bold(true).Render(fmt.Sprintf("%2d.", i+1))
                sb.WriteString(fmt.Sprintf("  %s %s\n", num, valueStyle.Render(item)))
        }
        return sb.String()
}

func Step(num int, label string) string {
        return fmt.Sprintf("  %s %s", accent.Bold(true).Render(fmt.Sprintf("%d.", num)), Bold(label))
}

func StepDone(num int, label string) string {
        return fmt.Sprintf("  %s %s", success.Bold(true).Render("✓"), success.Render(label))
}

func Hint(text string) string {
        return Dim("  → " + text)
}

func padRight(s string, width int) string {
        if len(s) >= width {
                return s
        }
        return s + strings.Repeat(" ", width-len(s))
}

// DashboardConfig holds all parameters for rendering the dashboard.
type DashboardConfig struct {
        ProjectDir  string
        HasDB       bool
        HasAPI      bool
        HasGRPC     bool
        HasLog      bool
        MissingDeps int
        Version     string
}

// Dashboard renders the main dashboard view.
func Dashboard(cfg DashboardConfig) string {
        var sb strings.Builder

        sb.WriteString(Banner(cfg.Version))
        sb.WriteString("\n\n")

        sb.WriteString("  ")
        sb.WriteString(labelStyle.Render("Project: "))
        sb.WriteString(valueStyle.Render(cfgDir(cfg.ProjectDir)))
        sb.WriteString("\n\n")

        sb.WriteString("  ")
        sb.WriteString(labelStyle.Render("Services"))
        sb.WriteString("\n")

        if cfg.HasDB {
                sb.WriteString("  ")
                sb.WriteString(DBLabel())
                sb.WriteString(" ")
                sb.WriteString(success.Render("Database"))
                sb.WriteString("\n")
        }
        if cfg.HasAPI {
                sb.WriteString("  ")
                sb.WriteString(APILabel())
                sb.WriteString(" ")
                sb.WriteString(success.Render("HTTP API"))
                sb.WriteString("\n")
        }
        if cfg.HasGRPC {
                sb.WriteString("  ")
                sb.WriteString(GRPCLabel())
                sb.WriteString(" ")
                sb.WriteString(success.Render("gRPC"))
                sb.WriteString("\n")
        }
        if cfg.HasLog {
                sb.WriteString("  ")
                sb.WriteString(LogLabel())
                sb.WriteString(" ")
                sb.WriteString(success.Render("Logs"))
                sb.WriteString("\n")
        }

        if !cfg.HasDB && !cfg.HasAPI && !cfg.HasGRPC && !cfg.HasLog {
                sb.WriteString(Dim("  No services configured"))
                sb.WriteString("\n")
        }

        sb.WriteString("\n")

        if cfg.MissingDeps > 0 {
                sb.WriteString("  ")
                sb.WriteString(warning.Render(fmt.Sprintf("⚠ %d missing dependencies", cfg.MissingDeps)))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker deps install"))
                sb.WriteString("\n\n")
        }

        sb.WriteString("  ")
        sb.WriteString(labelStyle.Render("Quick Commands"))
        sb.WriteString("\n")

        if cfg.HasDB {
                sb.WriteString(Hint("tinker db              Interactive database session"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker db ls           List tables"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker db desc users   Show table schema"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker db c users      Count rows"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker db f users 1    Find row by ID"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker db sql \"SELECT\" Run SQL"))
                sb.WriteString("\n")
        }
        if cfg.HasAPI {
                sb.WriteString(Hint("tinker api GET /users  Call an API endpoint"))
                sb.WriteString("\n")
        }
        if cfg.HasGRPC {
                sb.WriteString(Hint("tinker grpc list       List gRPC services"))
                sb.WriteString("\n")
        }
        if cfg.HasLog {
                sb.WriteString(Hint("tinker log             Show log file"))
                sb.WriteString("\n")
                sb.WriteString(Hint("tinker log tail        Follow log in real-time"))
                sb.WriteString("\n")
        }
        sb.WriteString(Hint("tinker config validate Check configuration"))
        sb.WriteString("\n")
        sb.WriteString(Hint("tinker make build      Run a Makefile target"))
        sb.WriteString("\n")

        return sb.String()
}

func cfgDir(path string) string {
        parts := strings.Split(path, "/")
        if len(parts) > 3 {
                return strings.Join(parts[len(parts)-3:], "/")
        }
        return path
}
