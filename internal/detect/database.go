package detect

import (
        "fmt"
        "os"
        "strings"
)

var dbEnvNames = []string{
        "DATABASE_URL", "DB_URL", "DB_HOST", "DB_CONNECTION",
        "POSTGRES_URL", "MYSQL_URL", "MONGO_URL", "MONGODB_URI",
        "DB_PATH", "SQLITE_PATH", "SQLITE_DB",
        "DB_DATABASE", "DB_NAME",
}

var dbFilePatterns = []string{".sqlite3", ".sqlite", ".db"}

func (d *Detector) detectDatabase() *DatabaseResult {
        var result *DatabaseResult

        // 1. Check for database env vars
        for _, name := range dbEnvNames {
                if v := d.getEnv(name); v != "" {
                        result = &DatabaseResult{
                                Source: fmt.Sprintf("env:%s", name),
                                Type:   inferDBType(d.getEnv, name),
                        }
                        break
                }
        }

        // 2. Scan for .db/.sqlite files in the project directory
        if result == nil {
                if path, ok := d.findDBFile(); ok {
                        result = &DatabaseResult{Source: path, Type: "sqlite3"}
                }
        }

        if result == nil {
                return nil
        }

        // 3. Detect migration and seed directories
        result.MigrateDir = d.findDir(migrateDirCandidates)
        result.SeedDir = d.findDir(seedDirCandidates)

        return result
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

        // URL-style connection strings
        switch {
        case strings.HasPrefix(v, "postgres://"), strings.HasPrefix(v, "postgresql://"):
                return "postgres"
        case strings.HasPrefix(v, "mysql://"), strings.HasPrefix(v, "mysql2://"):
                return "mysql"
        case strings.HasPrefix(v, "mongodb://"), strings.HasPrefix(v, "mongodb+srv://"):
                return "mongodb"
        }

        // File paths — SQLite
        if strings.HasSuffix(v, ".db") || strings.HasSuffix(v, ".sqlite") ||
                strings.HasSuffix(v, ".sqlite3") {
                return "sqlite3"
        }

        // Port-based inference
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

        // Env var name heuristics
        name := strings.ToUpper(primary)
        switch {
        case strings.Contains(name, "POSTGRES") || strings.Contains(name, "PG"):
                return "postgres"
        case strings.Contains(name, "MYSQL"):
                return "mysql"
        case strings.Contains(name, "MONGO"):
                return "mongodb"
        case strings.Contains(name, "SQLITE") || strings.Contains(name, "DB_PATH"):
                return "sqlite3"
        }

        return "unknown"
}

var migrateDirCandidates = []string{
        "migrations",
        "migrate",
        "db/migrations",
        "db/migrate",
        "sql/migrations",
        "backend/migrations",
        "backend/migrate",
}

var seedDirCandidates = []string{
        "seed",
        "seeds",
        "db/seed",
        "db/seeds",
        "sql/seed",
        "sql/seeds",
        "backend/seed",
        "backend/seeds",
}

// findDir returns the first candidate directory that exists, or "" if none found.
func (d *Detector) findDir(candidates []string) string {
        for _, c := range candidates {
                if info, err := os.Stat(d.dir + "/" + c); err == nil && info.IsDir() {
                        return c
                }
        }
        return ""
}
