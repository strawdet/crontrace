package cli

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/user/crontrace/internal/store"
)

// ManageBaseline dispatches baseline sub-commands: set, del, list.
func ManageBaseline(db *sql.DB, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: crontrace baseline <set|del|list> [args...]")
		os.Exit(1)
	}
	repo, err := store.NewBaselineRepository(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "baseline: %v\n", err)
		os.Exit(1)
	}
	manageBaseline(repo, args, os.Stdout)
}

func manageBaseline(repo *store.BaselineRepository, args []string, w io.Writer) {
	switch args[0] {
	case "set":
		baselineSet(repo, args[1:], w)
	case "del":
		baselineDel(repo, args[1:], w)
	case "list":
		baselineList(repo, w)
	default:
		fmt.Fprintf(w, "unknown baseline sub-command: %s\n", args[0])
	}
}

func baselineSet(repo *store.BaselineRepository, args []string, w io.Writer) {
	if len(args) < 2 {
		fmt.Fprintln(w, "usage: baseline set <command> <avg_ms> [sample_count]")
		return
	}
	avgMs, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Fprintf(w, "invalid avg_ms: %v\n", err)
		return
	}
	samples := 1
	if len(args) >= 3 {
		if n, err := strconv.Atoi(args[2]); err == nil {
			samples = n
		}
	}
	b := store.Baseline{
		Command:       args[0],
		AvgDurationMs: avgMs,
		SampleCount:   samples,
		UpdatedAt:     nowUTC(),
	}
	if err := repo.Upsert(b); err != nil {
		fmt.Fprintf(w, "set baseline: %v\n", err)
		return
	}
	fmt.Fprintf(w, "baseline set for %q: avg=%dms samples=%d\n", b.Command, b.AvgDurationMs, b.SampleCount)
}

func baselineDel(repo *store.BaselineRepository, args []string, w io.Writer) {
	if len(args) < 1 {
		fmt.Fprintln(w, "usage: baseline del <command>")
		return
	}
	if err := repo.Delete(args[0]); err != nil {
		fmt.Fprintf(w, "del baseline: %v\n", err)
		return
	}
	fmt.Fprintf(w, "baseline deleted for %q\n", args[0])
}

func baselineList(repo *store.BaselineRepository, w io.Writer) {
	list, err := repo.List()
	if err != nil {
		fmt.Fprintf(w, "list baselines: %v\n", err)
		return
	}
	if len(list) == 0 {
		fmt.Fprintln(w, "no baselines configured")
		return
	}
	fmt.Fprintf(w, "%-40s %10s %8s  %s\n", "COMMAND", "AVG_MS", "SAMPLES", "UPDATED")
	for _, b := range list {
		fmt.Fprintf(w, "%-40s %10d %8d  %s\n", b.Command, b.AvgDurationMs, b.SampleCount, b.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
}
