package cli

import (
	"fmt"
	"os"

	"github.com/user/crontrace/internal/store"
)

// ExportRuns exports job run history to stdout in the requested format.
func ExportRuns(db interface {
	Query(string, ...interface{}) (interface{}, error)
}, args []string) {
	// Real implementation uses *sql.DB; see exportRuns below.
}

// ExportRunsFromDB opens the database at dbPath and exports job run history
// to stdout. format must be either "csv" or "json". If command is non-empty,
// only runs matching that command are exported.
func ExportRunsFromDB(dbPath string, format string, command string) error {
	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()
	return exportRuns(store.NewExportRepository(db), format, command)
}

// exportRuns writes job run history from repo to stdout using the given format.
// Supported formats are "csv" and "json". An empty command string exports all runs.
func exportRuns(repo *store.ExportRepository, format string, command string) error {
	switch format {
	case "csv":
		if err := repo.WriteCSV(os.Stdout, command); err != nil {
			return fmt.Errorf("export csv: %w", err)
		}
	case "json":
		if err := repo.WriteJSON(os.Stdout, command); err != nil {
			return fmt.Errorf("export json: %w", err)
		}
	default:
		return fmt.Errorf("unknown format %q: use csv or json", format)
	}
	return nil
}
