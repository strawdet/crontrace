package cli

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/crontrace/internal/store"
)

// PrintDashboard prints an aggregated summary of all cron job runs.
func PrintDashboard(db *sql.DB) {
	repo := store.NewDashboardRepository(db)
	printDashboard(os.Stdout, repo)
}

func printDashboard(w io.Writer, repo *store.DashboardRepository) {
	s, err := repo.GetSummary()
	if err != nil {
		fmt.Fprintf(w, "error fetching dashboard: %v\n", err)
		return
	}

	fmt.Fprintln(w, "=== crontrace dashboard ===")
	fmt.Fprintf(w, "Total runs      : %d\n", s.TotalRuns)
	fmt.Fprintf(w, "Successful      : %d\n", s.SuccessRuns)
	fmt.Fprintf(w, "Failed          : %d\n", s.FailedRuns)
	fmt.Fprintf(w, "Unique commands : %d\n", s.UniqueCommands)

	if s.TotalRuns > 0 {
		fmt.Fprintf(w, "Avg duration    : %.0f ms\n", s.AvgDurationMs)
	} else {
		fmt.Fprintf(w, "Avg duration    : n/a\n")
	}

	if s.LastRunAt != nil {
		fmt.Fprintf(w, "Last run at     : %s\n", s.LastRunAt.Format(time.RFC3339))
	} else {
		fmt.Fprintf(w, "Last run at     : n/a\n")
	}
}
