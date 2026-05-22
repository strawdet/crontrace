package cli

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/user/crontrace/internal/store"
)

// PrintStats writes per-command aggregate statistics to stdout.
func PrintStats(dbPath string) error {
	return printStats(dbPath, os.Stdout)
}

func printStats(dbPath string, w io.Writer) error {
	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	repo := store.NewStatsRepository(db)
	stats, err := repo.ByCommand()
	if err != nil {
		return fmt.Errorf("query stats: %w", err)
	}

	if len(stats) == 0 {
		fmt.Fprintln(w, "No job runs recorded yet.")
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "COMMAND\tTOTAL\tSUCCESS\tFAILURE\tAVG DURATION (s)\tLAST EXIT")
	for _, s := range stats {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%.2f\t%d\n",
			s.Command,
			s.TotalRuns,
			s.SuccessRuns,
			s.FailureRuns,
			s.AvgDurationS,
			s.LastExitCode,
		)
	}
	return tw.Flush()
}
