package db

import (
        "fmt"

        "github.com/charmbracelet/bubbles/table"
        tea "github.com/charmbracelet/bubbletea"
)

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

                if len(msg.headers) == 0 {
                        m.err = fmt.Errorf("no columns returned for table")
                        return m, nil
                }

                availWidth := m.width
                if availWidth <= 0 {
                        availWidth = 80
                }

                columns := make([]table.Column, len(msg.headers))
                colWidth := availWidth / len(msg.headers)
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
