package detect

import (
	"fmt"
	"os"
	"strings"
)

var dbEnvNames = []string{
	"DATABASE_URL", "DB_URL", "DB_HOST", "DB_CONNECTION",
	"POSTGRES_URL", "MYSQL_URL", "MONGO_URL", "MONGODB_URI", "SQLITE_PATH",
}

var dbFilePatterns = []string{".sqlite3", ".sqlite", ".db"}

func (d *Detector) detectDatabase() *DatabaseResult {
	if path, ok := d.findDBFile(); ok {
		return &DatabaseResult{Source: path, Type: "sqlite3"}
	}

	for _, name := range dbEnvNames {
		if v := d.getEnv(name); v != "" {
			return &DatabaseResult{
				Source: fmt.Sprintf("env:%s", name),
				Type:   inferDBType(d.getEnv, name),
			}
		}
	}

	return nil
}

func (d *Detector) findDBFile() (string, bool) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return "", false
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		for _, ext := range dbFilePatterns {
			if strings.HasSuffix(strings.ToLower(e.Name()), ext) {
				return e.Name(), true
			}
		}
	}
	return "", false
}

func inferDBType(get func(string) string, primary string) string {
	v := strings.ToLower(get(primary))
	switch {
	case strings.HasPrefix(v, "postgres://"), strings.HasPrefix(v, "postgresql://"):
		return "postgres"
	case strings.HasPrefix(v, "mysql://"), strings.HasPrefix(v, "mysql2://"):
		return "mysql"
	case strings.HasPrefix(v, "mongodb://"), strings.HasPrefix(v, "mongodb+srv://"):
		return "mongodb"
	}

	port := get("DB_PORT")
	switch port {
	case "5432":
		return "postgres"
	case "3306":
		return "mysql"
	case "27017":
		return "mongodb"
	case "1433":
		return "sqlserver"
	}

	if strings.Contains(v, "5432") {
		return "postgres"
	}
	if strings.Contains(v, "3306") {
		return "mysql"
	}

	return "unknown"
}
