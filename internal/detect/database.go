package detect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var dbEnvNames = []string{
	"DATABASE_URL", "DB_URL", "DB_HOST",
	"POSTGRES_URL", "MYSQL_URL", "MONGO_URL", "MONGODB_URI", "SQLITE_PATH",
}

var dbFiles = []string{"db.sqlite3", "db.sqlite", "database.sqlite3", "test.db"}

func (d *Detector) detectDatabase() *DatabaseResult {
	for _, name := range dbEnvNames {
		if v := d.getEnv(name); v != "" {
			return &DatabaseResult{
				Source: fmt.Sprintf("env:%s", name),
				Type:   inferDBType(v),
			}
		}
	}

	for _, f := range dbFiles {
		if _, err := os.Stat(filepath.Join(d.dir, f)); err == nil {
			return &DatabaseResult{Source: f, Type: "sqlite3"}
		}
	}

	return nil
}

func inferDBType(dsn string) string {
	l := strings.ToLower(dsn)
	switch {
	case strings.HasPrefix(l, "postgres://"), strings.HasPrefix(l, "postgresql://"):
		return "postgres"
	case strings.HasPrefix(l, "mysql://"), strings.HasPrefix(l, "mysql2://"):
		return "mysql"
	case strings.HasPrefix(l, "mongodb://"), strings.HasPrefix(l, "mongodb+srv://"):
		return "mongodb"
	case strings.Contains(l, "5432"):
		return "postgres"
	case strings.Contains(l, "3306"):
		return "mysql"
	default:
		return "unknown"
	}
}
