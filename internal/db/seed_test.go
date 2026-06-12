package db

import "testing"

func TestSplitSQL(t *testing.T) {
        tests := []struct {
                name    string
                input   string
                want    []string
        }{
                {
                        name:  "empty content",
                        input: "",
                        want:  nil,
                },
                {
                        name:  "whitespace only",
                        input: "   \n\t  ",
                        want:  nil,
                },
                {
                        name:  "single statement with semicolon",
                        input: "SELECT 1;",
                        want:  []string{"SELECT 1"},
                },
                {
                        name:  "single statement without trailing semicolon",
                        input: "SELECT 1",
                        want:  []string{"SELECT 1"},
                },
                {
                        name:  "multiple statements separated by semicolons",
                        input: "CREATE TABLE users (id INT); INSERT INTO users VALUES (1);",
                        want:  []string{"CREATE TABLE users (id INT)", "INSERT INTO users VALUES (1)"},
                },
                {
                        name:  "semicolon inside single-quoted string should not split",
                        input: "INSERT INTO logs (msg) VALUES ('hello; world');",
                        want:  []string{"INSERT INTO logs (msg) VALUES ('hello; world')"},
                },
                {
                        name:  "escaped single quotes inside string",
                        input: "INSERT INTO users (name) VALUES ('O''Brien');",
                        want:  []string{"INSERT INTO users (name) VALUES ('O''Brien')"},
                },
                {
                        name:  "line comment is stripped",
                        input: "-- this is a comment\nSELECT 1;",
                        want:  []string{"SELECT 1"},
                },
                {
                        name:  "line comment in middle of statement",
                        input: "SELECT 1 -- comment\nFROM users;",
                        want:  []string{"SELECT 1 FROM users"},
                },
                {
                        name:  "multiple line comments",
                        input: "-- comment 1\n-- comment 2\nSELECT 1;",
                        want:  []string{"SELECT 1"},
                },
                {
                        name:  "comment inside quoted string is preserved",
                        input: "INSERT INTO t (v) VALUES ('-- not a comment');",
                        want:  []string{"INSERT INTO t (v) VALUES ('-- not a comment')"},
                },
                {
                        name:  "multiple semicolons with empty between",
                        input: "SELECT 1;; SELECT 2;",
                        want:  []string{"SELECT 1", "SELECT 2"},
                },
                {
                        name:  "trailing semicolons only",
                        input: ";;;",
                        want:  nil,
                },
                {
                        name:  "mixed statements and comments",
                        input: "-- setup\nCREATE TABLE t (id INT);\n-- data\nINSERT INTO t VALUES (1);",
                        want:  []string{"CREATE TABLE t (id INT)", "INSERT INTO t VALUES (1)"},
                },
                {
                        name:  "statement with multiple quoted strings containing semicolons",
                        input: "INSERT INTO t VALUES ('a;b', 'c;d');",
                        want:  []string{"INSERT INTO t VALUES ('a;b', 'c;d')"},
                },
                {
                        name:  "complex real-world seed",
                        input: "-- Create users table\nCREATE TABLE users (\n  id INTEGER PRIMARY KEY,\n  name TEXT NOT NULL\n);\n-- Seed data\nINSERT INTO users (name) VALUES ('Alice');\nINSERT INTO users (name) VALUES ('Bob; Not Charlie');\n",
                        want: []string{
                                "CREATE TABLE users (\n  id INTEGER PRIMARY KEY,\n  name TEXT NOT NULL\n)",
                                "INSERT INTO users (name) VALUES ('Alice')",
                                "INSERT INTO users (name) VALUES ('Bob; Not Charlie')",
                        },
                },
                {
                        name:  "comment at end of line without newline",
                        input: "SELECT 1 -- end comment",
                        want:  []string{"SELECT 1"},
                },
                {
                        name:  "statement ending without semicolon after comment",
                        input: "SELECT 1 -- trailing\n",
                        want:  []string{"SELECT 1"},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := splitSQL(tt.input)
                        if len(got) != len(tt.want) {
                                t.Errorf("splitSQL() returned %d statements, want %d\n  got:  %q\n  want: %q",
                                        len(got), len(tt.want), got, tt.want)
                                return
                        }
                        for i := range got {
                                if got[i] != tt.want[i] {
                                        t.Errorf("splitSQL()[%d] = %q, want %q", i, got[i], tt.want[i])
                                }
                        }
                })
        }
}
