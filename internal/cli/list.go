package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/crontrace/internal/store"
)

// ListOptions holds configuration for the list command.
type ListOptions struct {
	JobName string
	Limit   int
}

// ListRuns prints recent job runs from the store to stdout.
func ListRuns(repo *store.JobRunRepository, opts ListOptions) error {
	runs, err := repo.List(opts.JobName, opts.Limit)
	if err != nil {
		return fmt.Errorf("listing runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No job runs found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tJOB\tSTARTED\tDURATION\tEXIT CODE")
	fmt.Fprintln(w, "--\t---\t-------\t--------\t---------")

	for _, r := range runs {
		duration := "-"
		if r.FinishedAt != nil {
			d := r.FinishedAt.Sub(r.StartedAt).Round(time.Millisecond)
			duration = d.String()
		}

		exitCode := "-"
		if r.ExitCode != nil {
			exitCode = fmt.Sprintf("%d", *r.ExitCode)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			r.ID,
			r.JobName,
			r.StartedAt.Format(time.RFC3339),
			duration,
			exitCode,
		)
	}

	return w.Flush()
}
