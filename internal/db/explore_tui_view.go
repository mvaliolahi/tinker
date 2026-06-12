package db

import (
	"fmt"
	"strings"
)

// View renders the TUI.
func (m ExploreModel) View() string {
	var sb strings.Builder

	// Header
	sb.WriteString(exploreTitleStyle.Render(fmt.Sprintf(" DB Explorer — %s ", m.session.Type)))
	sb.WriteString("\n")

	switch m.view {
	case viewTables:
		sb.WriteString(m.listView.View())
		sb.WriteString("\n")
		sb.WriteString(exploreHelpStyle.Render("  ↑/k ↓/j navigate • enter: view data • s: schema • q: quit"))

	case viewData:
		sb.WriteString(exploreStatusBar.Render(fmt.Sprintf(" %s  •  %d rows (showing up to 100)", m.tableName, m.rowCount)))
		sb.WriteString("\n")
		sb.WriteString(m.tableView.View())
		sb.WriteString("\n")
		sb.WriteString(exploreHelpStyle.Render("  esc: back • s: schema • ↑/↓ scroll"))

	case viewSchema:
		sb.WriteString(exploreStatusBar.Render(fmt.Sprintf(" Schema: %s", m.tableName)))
		sb.WriteString("\n")
		sb.WriteString(m.viewport.View())
		sb.WriteString("\n")
		sb.WriteString(exploreHelpStyle.Render("  esc: back • ↑/↓ scroll"))
	}

	if m.loading {
		sb.WriteString("\n  Loading...")
	}
	if m.err != nil {
		sb.WriteString("\n" + exploreErrorStyle.Render(fmt.Sprintf("  Error: %s", m.err.Error())))
	}

	return sb.String()
}

// tableNamesForExplore returns table names for the explorer.
func (s *Session) tableNamesForExplore() ([]string, error) {
	if s.db != nil {
		return s.tableNamesNative()
	}
	// CLI fallback
	out, err := s.tablesCLI()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			names = append(names, t)
		}
	}
	return names, nil
}
