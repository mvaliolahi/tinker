package db

import (
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
