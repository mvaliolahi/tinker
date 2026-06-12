package db

import "testing"

func TestQuoteIdent(t *testing.T) {
        tests := []struct {
                name  string
                input string
                want  string
        }{
                {"simple name", "users", `"users"`},
                {"already double-quoted", `"users"`, `"users"`},
                {"already backtick-quoted", "`users`", `"users"`},
                {"double-quoted with inner spaces", `"my table"`, `"my table"`},
                {"backtick-quoted with inner spaces", "`my table`", `"my table"`},
                {"empty string", "", `""`},
                {"name with schema", "public.users", `"public.users"`},
                {"name with special chars", "tbl$123", `"tbl$123"`},
                {"only double quotes", `""`, `""`},
                {"only backticks", "``", `""`},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := quoteIdent(tt.input)
                        if got != tt.want {
                                t.Errorf("quoteIdent(%q) = %q, want %q", tt.input, got, tt.want)
                        }
                })
        }
}

func TestQuoteStr(t *testing.T) {
        tests := []struct {
                name  string
                input string
                want  string
        }{
                {"simple name", "users", `'users'`},
                {"already single-quoted", "'users'", `'users'`},
                {"already double-quoted", `"users"`, `'users'`},
                {"name with single quote", "user's", `'user''s'`},
                {"empty string", "", `''`},
                {"name with multiple single quotes", "a'b'c", `'a''b''c'`},
                {"double-quoted with spaces", `"my table"`, `'my table'`},
                {"only single quotes", "''", `''`}, // '' is stripped to empty, then wrapped: ''
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := quoteStr(tt.input)
                        if got != tt.want {
                                t.Errorf("quoteStr(%q) = %q, want %q", tt.input, got, tt.want)
                        }
                })
        }
}

func TestEscapeSingle(t *testing.T) {
        tests := []struct {
                name  string
                input string
                want  string
        }{
                {"no single quotes", "hello", "hello"},
                {"one single quote", "it's", "it''s"},
                {"multiple single quotes", "it's a test's", "it''s a test''s"},
                {"consecutive single quotes", "a''b", "a''''b"},
                {"empty string", "", ""},
                {"only single quote", "'", "''"},
                {"single quote at start", "'hello", "''hello"},
                {"single quote at end", "hello'", "hello''"},
                {"no letters just quote", "'", "''"},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := escapeSingle(tt.input)
                        if got != tt.want {
                                t.Errorf("escapeSingle(%q) = %q, want %q", tt.input, got, tt.want)
                        }
                })
        }
}
