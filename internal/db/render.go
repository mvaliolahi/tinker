package db

import (
        "database/sql"
        "fmt"
        "strings"

        "github.com/jedib0t/go-pretty/v6/table"
        "github.com/jedib0t/go-pretty/v6/text"
)

// renderTable renders database query results as a formatted table using go-pretty.
// It takes column names and rows from a sql.Rows result set and returns
// a beautifully formatted string with aligned columns, borders, and a header row.
func renderTable(cols []string, rows [][]string) string {
        if len(rows) == 0 && len(cols) == 0 {
                return ""
        }

        t := table.NewWriter()
        t.SetStyle(tableStyle())

        // Add header row
        headers := make(table.Row, len(cols))
        for i, col := range cols {
                headers[i] = col
        }
        t.AppendHeader(headers)

        // Add data rows
        for _, row := range rows {
                cells := make(table.Row, len(row))
                for i, cell := range row {
                        cells[i] = cell
                }
                t.AppendRow(cells)
        }

        return t.Render() + "\n"
}

// renderTableFromRows renders a sql.Rows result set as a formatted table.
func renderTableFromRows(rows *sql.Rows) (string, error) {
        cols, err := rows.Columns()
        if err != nil {
                return "", err
        }

        var dataRows [][]string
        for rows.Next() {
                vals := make([]interface{}, len(cols))
                ptrs := make([]interface{}, len(cols))
                for i := range vals {
                        ptrs[i] = &vals[i]
                }
                if err := rows.Scan(ptrs...); err != nil {
                        continue
                }

                strVals := make([]string, len(cols))
                for i, v := range vals {
                        switch val := v.(type) {
                        case []byte:
                                strVals[i] = string(val)
                        case string:
                                strVals[i] = val
                        case nil:
                                strVals[i] = "NULL"
                        default:
                                strVals[i] = fmt.Sprintf("%v", val)
                        }
                }
                dataRows = append(dataRows, strVals)
        }

        return renderTable(cols, dataRows), rows.Err()
}

// renderSimpleList renders a single-column result as a simple list
// (for commands like `tables` that return just names).
func renderSimpleList(items []string) string {
        var sb strings.Builder
        for _, item := range items {
                sb.WriteString(item)
                sb.WriteString("\n")
        }
        return sb.String()
}

// tableStyle returns a tinker-themed table style that matches our lipgloss aesthetic.
func tableStyle() table.Style {
        return table.Style{
                Name: "tinker",
                Box: table.BoxStyle{
                        BottomLeft:       "",
                        BottomRight:      "",
                        BottomSeparator:  "─",
                        Left:             "",
                        LeftSeparator:    "",
                        MiddleHorizontal: "─",
                        MiddleSeparator:  "─",
                        MiddleVertical:   "  ",
                        PaddingLeft:      "",
                        PaddingRight:     "  ",
                        PageSeparator:    "\n",
                        Right:            "",
                        RightSeparator:   "",
                        TopLeft:          "",
                        TopRight:         "",
                        TopSeparator:     "",
                        UnfinishedRow:    " ~",
                },
                Color: table.ColorOptions{
                        Header:       text.Colors{text.FgHiCyan, text.Bold},
                        Row:          text.Colors{text.FgHiWhite},
                        RowAlternate: text.Colors{text.FgWhite},
                        Footer:       text.Colors{text.FgHiYellow, text.Bold},
                        Separator:    text.Colors{text.FgHiBlack},
                },
                Format: table.FormatOptions{
                        Footer: text.FormatUpper,
                        Header: text.FormatUpper,
                        Row:    text.FormatDefault,
                },
                Options: table.Options{
                        DrawBorder:      false,
                        SeparateColumns: false,
                        SeparateFooter:  true,
                        SeparateHeader:  true,
                        SeparateRows:    false,
                },
        }
}
