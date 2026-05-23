package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/user/crontrace/internal/runner"
	"github.com/user/crontrace/internal/store"
)

// ReplayRun re-executes a previous job run identified by its ID, or lists
// replayable commands when called with no arguments.
func ReplayRun(db *sql.DB, args []string) error {
	repo := store.NewReplayRepository(db)
	jr := store.NewJobRunRepository(db)

	if len(args) == 0 {
		return listReplayable(repo)
	}

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid run id %q: %w", args[0], err)
	}

	entry, err := repo.GetReplayByID(id)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Replaying run #%d: %s %s\n", entry.ID, entry.Command, entry.Args)

	cmdArgs := []string{}
	if entry.Args != "" {
		cmdArgs = append(cmdArgs, entry.Args)
	}

	_, runErr := runner.Run(jr, entry.Command, cmdArgs)
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "replay finished with error: %v\n", runErr)
	}
	return nil
}

func listReplayable(repo *store.ReplayRepository) error {
	entries, err := repo.ListReplayable(20)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No replayable runs found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCOMMAND\tARGS\tLAST RUN\tEXIT CODE")
	for _, e := range entries {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n",
			e.ID,
			e.Command,
			e.Args,
			e.LastRunAt.Format("2006-01-02 15:04:05"),
			e.ExitCode,
		)
	}
	return w.Flush()
}
