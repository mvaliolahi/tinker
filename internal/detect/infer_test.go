package detect

import (
	"testing"
)

func TestInferDBType_URLStyle(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		envVal  string
		want    string
	}{
		{"postgres URL", "DATABASE_URL", "postgres://user:pass@host:5432/db", "postgres"},
		{"postgresql URL", "DATABASE_URL", "postgresql://host/db", "postgres"},
		{"mysql URL", "DATABASE_URL", "mysql://host/db", "mysql"},
		{"mysql2 URL", "DATABASE_URL", "mysql2://host/db", "mysql"},
		{"mongodb URL", "DATABASE_URL", "mongodb://host/db", "mongodb"},
		{"mongodb+srv URL", "DATABASE_URL", "mongodb+srv://host/db", "mongodb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := func(key string) string {
				if key == tt.envName {
					return tt.envVal
				}
				return ""
			}
			got := inferDBType(get, tt.envName)
			if got != tt.want {
				t.Errorf("inferDBType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferDBType_FilePath(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		envVal  string
		want    string
	}{
		{".db extension", "DB_PATH", "/path/to/file.db", "sqlite3"},
		{".sqlite extension", "SQLITE_PATH", "/path/to/file.sqlite", "sqlite3"},
		{".sqlite3 extension", "SQLITE_DB", "/path/to/file.sqlite3", "sqlite3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := func(key string) string {
				if key == tt.envName {
					return tt.envVal
				}
				return ""
			}
			got := inferDBType(get, tt.envName)
			if got != tt.want {
				t.Errorf("inferDBType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferDBType_PortBased(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		envVal  string
		dbPort  string
		want    string
	}{
		{"port 5432 -> postgres", "DATABASE_URL", "somehost", "5432", "postgres"},
		{"port 3306 -> mysql", "DATABASE_URL", "somehost", "3306", "mysql"},
		{"port 27017 -> mongodb", "DATABASE_URL", "somehost", "27017", "mongodb"},
		{"port 1433 -> sqlserver", "DATABASE_URL", "somehost", "1433", "sqlserver"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := func(key string) string {
				if key == tt.envName {
					return tt.envVal
				}
				if key == "DB_PORT" {
					return tt.dbPort
				}
				return ""
			}
			got := inferDBType(get, tt.envName)
			if got != tt.want {
				t.Errorf("inferDBType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferDBType_EnvNameHeuristics(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		want    string
	}{
		{"POSTGRES_URL", "POSTGRES_URL", "postgres"},
		{"PG_HOST", "PG_HOST", "postgres"},
		{"MYSQL_URL", "MYSQL_URL", "mysql"},
		{"MONGO_URI", "MONGO_URI", "mongodb"},
		{"SQLITE_PATH", "SQLITE_PATH", "sqlite3"},
		{"DB_PATH", "DB_PATH", "sqlite3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Return a non-empty value that doesn't match any URL or file pattern
			get := func(key string) string {
				if key == tt.envName {
					return "somevalue"
				}
				return ""
			}
			got := inferDBType(get, tt.envName)
			if got != tt.want {
				t.Errorf("inferDBType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInferDBType_UnknownFallback(t *testing.T) {
	get := func(key string) string {
		if key == "RANDOM_VAR" {
			return "something"
		}
		return ""
	}
	got := inferDBType(get, "RANDOM_VAR")
	if got != "unknown" {
		t.Errorf("inferDBType() = %q, want %q", got, "unknown")
	}
}

func TestInferDBType_PortInValue(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		envVal  string
		want    string
	}{
		{"port 5432 in value", "DATABASE_URL", "host:5432/mydb", "postgres"},
		{"port 3306 in value", "DATABASE_URL", "host:3306/mydb", "mysql"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := func(key string) string {
				if key == tt.envName {
					return tt.envVal
				}
				return ""
			}
			got := inferDBType(get, tt.envName)
			if got != tt.want {
				t.Errorf("inferDBType() = %q, want %q", got, tt.want)
			}
		})
	}
}
