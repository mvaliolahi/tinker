package ui

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/chroma/v2"
	chromaFormatters "github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// Highlight applies syntax highlighting to source code.
// language should be a Chroma lexer name: "sql", "json", "yaml", "go", "toml", etc.
// Falls back to plain text if highlighting fails.
func Highlight(source, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Use a dark style that looks good in terminals
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := chromaFormatters.Get("terminal256")
	if formatter == nil {
		formatter = chromaFormatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return source
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return source
	}

	return buf.String()
}

// HighlightSQL highlights a SQL statement.
func HighlightSQL(sql string) string {
	return Highlight(sql, "sql")
}

// HighlightJSON highlights a JSON string.
func HighlightJSON(json string) string {
	return Highlight(json, "json")
}

// HighlightYAML highlights a YAML string.
func HighlightYAML(yaml string) string {
	return Highlight(yaml, "yaml")
}

// HighlightGo highlights Go source code.
func HighlightGo(code string) string {
	return Highlight(code, "go")
}

// HighlightTOML highlights a TOML string.
func HighlightTOML(toml string) string {
	return Highlight(toml, "toml")
}

// FormatCount returns a styled count value.
func FormatCount(count int64) string {
	return fmt.Sprintf("%d\n", count)
}
