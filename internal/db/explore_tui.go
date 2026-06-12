package db

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// exploreView is the current view state of the explorer.
type exploreView int

const (
	viewTables exploreView = iota
	viewData
	viewSchema
)

// ExploreModel is the Bubble Tea model for the database explorer TUI.
type ExploreModel struct {
	session *Session
	width   int
	height  int

	view      exploreView
	tableView table.Model
	listView  list.Model
	viewport  viewport.Model

	tableName string
	schema    string
	rowCount  int64
	err       error
	loading   bool
}

var (
	exploreTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Padding(0, 1)
	exploreHelpStyle  = lipgloss.NewStyle().Faint(true).Padding(0, 1)
	exploreErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Padding(0, 1)
	exploreStatusBar  = lipgloss.NewStyle().Background(lipgloss.Color("23")).Foreground(lipgloss.Color("6")).Padding(0, 1)
)

// NewExploreModel creates a new database explorer TUI model.
func NewExploreModel(s *Session) ExploreModel {
	// Table list
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("12")).Bold(true)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(lipgloss.Color("15"))

	l := list.New([]list.Item{}, delegate, 80, 20)
	l.Title = "Tables"
	l.Styles.Title = exploreTitleStyle
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("up", "k")),
		CursorDown: key.NewBinding(key.WithKeys("down", "j")),
		Quit:       key.NewBinding(key.WithKeys("q")),
	}

	// Data table
	t := table.New()
	t.Focus()

	vp := viewport.New(80, 20)

	return ExploreModel{
		session:   s,
		listView:  l,
		tableView: t,
		viewport:  vp,
		view:      viewTables,
	}
}

// Init loads the table list.
func (m ExploreModel) Init() tea.Cmd {
	return m.loadTables
}

func (m ExploreModel) loadTables() tea.Msg {
	var tables []list.Item
	names, _ := m.session.tableNamesForExplore()

	for _, n := range names {
		tables = append(tables, tableItem{name: n})
	}
	return tablesLoadedMsg{tables: tables}
}

type tablesLoadedMsg struct {
	tables []list.Item
}

type dataLoadedMsg struct {
	headers []string
	rows    [][]string
	count   int64
}

type schemaLoadedMsg struct {
	schema string
}

type errorMsg struct {
	err error
}

// tableItem implements list.Item
type tableItem struct {
	name string
}

func (t tableItem) FilterValue() string { return t.name }
func (t tableItem) Title() string       { return t.name }
func (t tableItem) Description() string { return "" }

// Update handles Bubble Tea messages.
func (m ExploreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.listView.SetSize(msg.Width, msg.Height-4)
		m.tableView.SetWidth(msg.Width)
		m.tableView.SetHeight(msg.Height - 6)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
		return m, nil

	case tablesLoadedMsg:
		m.listView.SetItems(msg.tables)
		return m, nil

	case dataLoadedMsg:
		m.loading = false
		m.rowCount = msg.count

		columns := make([]table.Column, len(msg.headers))
		colWidth := m.width / len(msg.headers)
		if colWidth < 10 {
			colWidth = 10
		}
		for i, h := range msg.headers {
			w := colWidth
			if len(h) > w {
				w = len(h) + 2
			}
			columns[i] = table.Column{Title: h, Width: w}
		}

		rows := make([]table.Row, len(msg.rows))
		for i, r := range msg.rows {
			row := make(table.Row, len(r))
			for j, v := range r {
				if len(v) > colWidth-2 {
					v = v[:colWidth-3] + "…"
				}
				row[j] = v
			}
			rows[i] = row
		}

		m.tableView.SetColumns(columns)
		m.tableView.SetRows(rows)
		m.view = viewData
		return m, nil

	case schemaLoadedMsg:
		m.loading = false
		m.schema = msg.schema
		m.viewport.SetContent(m.schema)
		m.view = viewSchema
		return m, nil

	case errorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.loading {
			return m, nil
		}

		switch m.view {
		case viewTables:
			return m.handleTablesKey(msg)
		case viewData:
			return m.handleDataKey(msg)
		case viewSchema:
			return m.handleSchemaKey(msg)
		}
	}

	var cmd tea.Cmd
	switch m.view {
	case viewTables:
		m.listView, cmd = m.listView.Update(msg)
	case viewData:
		m.tableView, cmd = m.tableView.Update(msg)
	case viewSchema:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m ExploreModel) handleTablesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "enter":
		item, ok := m.listView.SelectedItem().(tableItem)
		if !ok {
			return m, nil
		}
		m.tableName = item.name
		m.loading = true
		m.err = nil
		return m, m.loadData(item.name)
	case "s":
		item, ok := m.listView.SelectedItem().(tableItem)
		if !ok {
			return m, nil
		}
		m.tableName = item.name
		m.loading = true
		m.err = nil
		return m, m.loadSchema(item.name)
	}

	var cmd tea.Cmd
	m.listView, cmd = m.listView.Update(msg)
	return m, cmd
}

func (m ExploreModel) handleDataKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.view = viewTables
		m.err = nil
		return m, nil
	case "s":
		m.loading = true
		return m, m.loadSchema(m.tableName)
	}

	var cmd tea.Cmd
	m.tableView, cmd = m.tableView.Update(msg)
	return m, cmd
}

func (m ExploreModel) handleSchemaKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.view = viewTables
		m.err = nil
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ExploreModel) loadData(table string) tea.Cmd {
	return func() tea.Msg {
		headers, rows, err := m.session.ExploreTable(table, 100)
		if err != nil {
			return errorMsg{err: err}
		}
		count, _ := m.session.ExploreRowCount(table)
		return dataLoadedMsg{headers: headers, rows: rows, count: count}
	}
}

func (m ExploreModel) loadSchema(table string) tea.Cmd {
	return func() tea.Msg {
		schema, err := m.session.Schema(table)
		if err != nil {
			return errorMsg{err: err}
		}
		return schemaLoadedMsg{schema: schema}
	}
}

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
