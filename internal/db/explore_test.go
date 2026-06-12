package db

import (
        "reflect"
        "testing"
)

func TestParseCLIOutput(t *testing.T) {
        tests := []struct {
                name        string
                input       string
                wantHeaders []string
                wantRows    [][]string
        }{
                {
                        name:        "empty output",
                        input:       "",
                        wantHeaders: nil,
                        wantRows:    nil,
                },
                {
                        name:        "whitespace only",
                        input:       "   \n   \n   ",
                        wantHeaders: nil,
                        wantRows:    nil,
                },
                {
                        name:        "single row header only",
                        input:       "id | name | email",
                        wantHeaders: []string{"id", "name", "email"},
                        wantRows:    nil,
                },
                {
                        name:        "header and separator only",
                        input:       "id | name\n---------",
                        wantHeaders: []string{"id", "name"},
                        wantRows:    nil,
                },
                {
                        name:        "header separator and one data row",
                        input:       "id | name\n---------\n1 | Alice",
                        wantHeaders: []string{"id", "name"},
                        wantRows:    [][]string{{"1", "Alice"}},
                },
                {
                        name:        "header separator and multiple data rows",
                        input:       "id | name | email\n-----------\n1 | Alice | alice@example.com\n2 | Bob | bob@example.com",
                        wantHeaders: []string{"id", "name", "email"},
                        wantRows: [][]string{
                                {"1", "Alice", "alice@example.com"},
                                {"2", "Bob", "bob@example.com"},
                        },
                },
                {
                        name:        "separator lines starting with dash are filtered",
                        input:       "id | name\n---\n1 | Alice\n------\n2 | Bob",
                        wantHeaders: []string{"id", "name"},
                        wantRows: [][]string{
                                {"1", "Alice"},
                                {"2", "Bob"},
                        },
                },
                {
                        name:        "single column output",
                        input:       "tables\n------\nusers\nposts\ncomments",
                        wantHeaders: []string{"tables"},
                        wantRows: [][]string{
                                {"users"},
                                {"posts"},
                                {"comments"},
                        },
                },
                {
                        name:        "pipe at edges with dash separator filtered",
                        input:       "id | name\n-----------\n1 | Alice",
                        wantHeaders: []string{"id", "name"},
                        wantRows:    [][]string{{"1", "Alice"}},
                },
                {
                        name:        "pipe at edges includes non-dash-starting separator as data",
                        input:       "| id | name |\n|----|------|\n| 1 | Alice |",
                        wantHeaders: []string{"", "id", "name", ""},
                        wantRows:    [][]string{{"", "----", "------", ""}, {"", "1", "Alice", ""}},
                },
                {
                        name:        "trailing newlines",
                        input:       "id | name\n--------\n1 | Alice\n\n\n",
                        wantHeaders: []string{"id", "name"},
                        wantRows:    [][]string{{"1", "Alice"}},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        gotHeaders, gotRows := parseCLIOutput(tt.input)
                        if !reflect.DeepEqual(gotHeaders, tt.wantHeaders) {
                                t.Errorf("parseCLIOutput() headers = %v, want %v", gotHeaders, tt.wantHeaders)
                        }
                        if !reflect.DeepEqual(gotRows, tt.wantRows) {
                                t.Errorf("parseCLIOutput() rows = %v, want %v", gotRows, tt.wantRows)
                        }
                })
        }
}

func TestSplitRow(t *testing.T) {
        tests := []struct {
                name  string
                input string
                want  []string
        }{
                {
                        name:  "simple two columns",
                        input: "id | name",
                        want:  []string{"id", "name"},
                },
                {
                        name:  "three columns",
                        input: "1 | Alice | alice@example.com",
                        want:  []string{"1", "Alice", "alice@example.com"},
                },
                {
                        name:  "single column no pipe",
                        input: "users",
                        want:  []string{"users"},
                },
                {
                        name:  "pipe with spaces",
                        input: "  id  |  name  ",
                        want:  []string{"id", "name"},
                },
                {
                        name:  "pipe at edges with spaces",
                        input: "| id | name |",
                        want:  []string{"", "id", "name", ""},
                },
                {
                        name:  "empty string",
                        input: "",
                        want:  []string{""},
                },
                {
                        name:  "only pipes",
                        input: "|||",
                        want:  []string{"", "", "", ""},
                },
                {
                        name:  "column values with spaces",
                        input: "1 | hello world | foo bar baz",
                        want:  []string{"1", "hello world", "foo bar baz"},
                },
                {
                        name:  "consecutive pipes with no content",
                        input: "a||b",
                        want:  []string{"a", "", "b"},
                },
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        got := splitRow(tt.input)
                        if !reflect.DeepEqual(got, tt.want) {
                                t.Errorf("splitRow(%q) = %v, want %v", tt.input, got, tt.want)
                        }
                })
        }
}
