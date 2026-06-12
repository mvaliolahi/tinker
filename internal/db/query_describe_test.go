package db

import "testing"

func TestValidateTable(t *testing.T) {
	// validateTable is a method on Session, but for testing pure validation
	// logic we can use a Session with no native connection (db == nil).
	// This skips the database existence check and only tests the input validation.
	s := &Session{Type: "sqlite3"}

	tests := []struct {
		name    string
		table   string
		wantErr bool
	}{
		// Valid table names
		{"simple name", "users", false},
		{"name with underscores", "user_accounts", false},
		{"name with digits", "tbl123", false},
		{"uppercase name", "Users", false},
		{"mixed case", "UserAccounts", false},
		{"name with dollar sign", "tbl$123", false},

		// Invalid table names — SQL injection patterns
		{"semicolon injection", "users; DROP TABLE users", true},
		{"double dash comment", "users--", true},
		{"block comment open", "users/*", true},
		{"block comment with content", "users/* comment */", true},
		{"space in name", "user accounts", true},
		{"tab in name", "user\taccounts", true},
		{"newline in name", "user\naccounts", true},
		{"carriage return", "user\raccounts", true},
		{"semicolon only", ";", true},
		{"comment prefix only", "--", true},
		{"block comment only", "/*", true},

		// Combined injection patterns
		{"union injection", "users; SELECT * FROM passwords", true},
		{"comment after name", "users -- drop", true},
		{"block comment injection", "users/* DROP TABLE users */", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateTable(tt.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTable(%q) error = %v, wantErr %v", tt.table, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTable_NoNativeConnection(t *testing.T) {
	// When there's no native connection, validateTable only does input validation.
	// A clean table name should pass.
	s := &Session{Type: "sqlite3"}

	if err := s.validateTable("clean_table"); err != nil {
		t.Errorf("validateTable with no native connection should pass for clean names, got: %v", err)
	}
}

func TestValidateTable_RejectsSemicolon(t *testing.T) {
	s := &Session{Type: "sqlite3"}
	err := s.validateTable("users; DROP TABLE users;")
	if err == nil {
		t.Error("expected error for table name with semicolon, got nil")
	}
}

func TestValidateTable_RejectsDoubleDash(t *testing.T) {
	s := &Session{Type: "sqlite3"}
	err := s.validateTable("users--")
	if err == nil {
		t.Error("expected error for table name with --, got nil")
	}
}

func TestValidateTable_RejectsBlockComment(t *testing.T) {
	s := &Session{Type: "sqlite3"}
	err := s.validateTable("users/*")
	if err == nil {
		t.Error("expected error for table name with /*, got nil")
	}
}

func TestValidateTable_RejectsWhitespace(t *testing.T) {
	s := &Session{Type: "sqlite3"}

	whitespaceNames := []string{"user name", "user\tname", "user\nname", "user\rname"}
	for _, name := range whitespaceNames {
		err := s.validateTable(name)
		if err == nil {
			t.Errorf("expected error for table name %q with whitespace, got nil", name)
		}
	}
}
