package cli

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/crontrace/internal/store"
)

// PruneRuns removes job run records based on age or command filter.
// olderThan: if > 0, delete runs older than this duration.
// command: if non-empty, delete all runs for this command.
func PruneRuns(db *sql.DB, olderThan time.Duration, command string) error {
	return pruneRuns(os.Stdout, db, olderThan, command)
}

func pruneRuns(w io.Writer, db *sql.DB, olderThan time.Duration, command string) error {
	repo := store.NewPruneRepository(db)

	if command != "" {
		n, err := repo.PruneByCommand(command)
		if err != nil {
			return fmt.Errorf("prune by command: %w", err)
		}
		fmt.Fprintf(w, "Pruned %d run(s) for command %q.\n", n, command)
		return nil
	}

	if olderThan <= 0 {
		return fmt.Errorf("must specify --older-than duration or --command")
	}

	n, err := repo.PruneOlderThan(olderThan)
	if err != nil {
		return fmt.Errorf("prune older than: %w", err)
	}
	fmt.Fprintf(w, "Pruned %d run(s) older than %s.\n", n, olderThan)
	return nil
}
