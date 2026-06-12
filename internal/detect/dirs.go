package detect

// MigrateDirCandidates lists common migration directory names (relative to project root).
// Shared between detect, contract, and CLI packages so the list is defined once.
var MigrateDirCandidates = []string{
	"migrations",
	"migrate",
	"db/migrations",
	"db/migrate",
	"sql/migrations",
	"backend/migrations",
	"backend/migrate",
}

// SeedDirCandidates lists common seed directory names (relative to project root).
var SeedDirCandidates = []string{
	"seed",
	"seeds",
	"db/seed",
	"db/seeds",
	"sql/seed",
	"sql/seeds",
	"backend/seed",
	"backend/seeds",
}
