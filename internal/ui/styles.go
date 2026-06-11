package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	bold     = lipgloss.NewStyle().Bold(true)
	faint    = lipgloss.NewStyle().Faint(true)
	accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	success  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	error_   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	warning  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	header   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
	dimLabel = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))

	sectionHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	badge = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	dbBadge   = badge.Foreground(lipgloss.Color("6")).Background(lipgloss.Color("23"))
	apiBadge  = badge.Foreground(lipgloss.Color("3")).Background(lipgloss.Color("58"))
	grpcBadge = badge.Foreground(lipgloss.Color("5")).Background(lipgloss.Color("53"))
	makeBadge = badge.Foreground(lipgloss.Color("2")).Background(lipgloss.Color("22"))

	indent = "  "
)

func Header(text string) string {
	return sectionHeader.Render(text)
}

func KeyValue(label, value string) string {
	return indent + labelStyle.Render(label+":") + " " + valueStyle.Render(value)
}

func Success(text string) string {
	return success.Bold(true).Render("✓") + " " + success.Render(text)
}

func Error(text string) string {
	return error_.Bold(true).Render("✗") + " " + error_.Render(text)
}

func Warning(text string) string {
	return warning.Bold(true).Render("⚠") + " " + warning.Render(text)
}

func Dim(text string) string {
	return faint.Render(text)
}

func Bold(text string) string {
	return bold.Render(text)
}

func Accent(text string) string {
	return accent.Render(text)
}

func DBLabel() string   { return dbBadge.Render(" DB ") }
func APILabel() string  { return apiBadge.Render(" API ") }
func GRPCLabel() string { return grpcBadge.Render(" GRPC ") }
func MakeLabel() string { return makeBadge.Render(" MAKE ") }

func Section(title string) string {
	line := strings.Repeat("─", 40)
	return "\n" + accent.Render(line) + "\n" + header.Render(title) + "\n" + accent.Render(line)
}

func Bullet(label, value string) string {
	return indent + accent.Render("▪") + " " + labelStyle.Render(label) + " " + valueStyle.Render(value)
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
	status := Success("installed")
	if !installed {
		status = Error("missing")
	}
	return fmt.Sprintf("%s%-10s %s  %s", indent, valueStyle.Render(name), status, labelStyle.Render(purpose))
}
