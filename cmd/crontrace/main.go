package main

import (
	"fmt"
	"os"

	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/store"
)

const defaultDBPath = ".crontrace.db"

func main() {
	dbPath := os.Getenv("CRONTRACE_DB")
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: crontrace run <command> [args...]")
			os.Exit(1)
		}
		repo, err := store.NewJobRunRepository(db)
		if err != nil {
			fatalf("job run repo: %v", err)
		}
		code, err := cli.Run(repo, os.Args[2], os.Args[3:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(code)
	case "list":
		if err := cli.ListRuns(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "stats":
		if err := cli.PrintStats(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "prune":
		if err := cli.PruneRuns(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "export":
		if err := cli.ExportRuns(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "tag":
		if err := cli.ManageTags(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "notify":
		if err := cli.ManageNotify(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	case "alert":
		if err := cli.ManageAlerts(db, os.Args[2:]); err != nil {
			fatalf("%v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `crontrace — cron job execution recorder

Usage:
  crontrace run <command> [args...]   Run a command and record its execution
  crontrace list [--cmd=<cmd>]        List recorded job runs
  crontrace stats [--cmd=<cmd>]       Show aggregated statistics
  crontrace prune --older-than=<d>    Delete old records
  crontrace export [--format=csv|json] Export records
  crontrace tag <set|del|list> ...    Manage tags on job runs
  crontrace notify <set|del|list> ... Manage notification rules
  crontrace alert <set|del|list> ...  Manage threshold-based alert rules`)
}
