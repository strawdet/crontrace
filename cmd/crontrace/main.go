package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/user/crontrace/internal/cli"
	"github.com/user/crontrace/internal/runner"
	"github.com/user/crontrace/internal/store"
)

const defaultDBPath = ".crontrace.db"

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}
	defaultDB := filepath.Join(homedir, defaultDBPath)

	dbPath := flag.String("db", defaultDB, "path to SQLite database")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: crontrace [options] <subcommand> [args]")
		fmt.Fprintln(os.Stderr, "subcommands: run, list, stats, prune")
		os.Exit(1)
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	switch args[0] {
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: crontrace run <command> [args...]")
			os.Exit(1)
		}
		repo := store.NewJobRunRepository(db)
		exitCode, runErr := runner.Run(repo, args[1], args[2:]...)
		if runErr != nil {
			fmt.Fprintf(os.Stderr, "run error: %v\n", runErr)
		}
		os.Exit(exitCode)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		limit := listCmd.Int("n", 20, "max number of runs to show")
		command := listCmd.String("command", "", "filter by command")
		listCmd.Parse(args[1:])
		if err := cli.ListRuns(db, *command, *limit); err != nil {
			log.Fatalf("list: %v", err)
		}

	case "stats":
		if err := cli.PrintStats(db); err != nil {
			log.Fatalf("stats: %v", err)
		}

	case "prune":
		pruneCmd := flag.NewFlagSet("prune", flag.ExitOnError)
		olderThanStr := pruneCmd.String("older-than", "", "delete runs older than this duration (e.g. 720h)")
		command := pruneCmd.String("command", "", "delete all runs for this command")
		pruneCmd.Parse(args[1:])

		var olderThan time.Duration
		if *olderThanStr != "" {
			olderThan, err = time.ParseDuration(*olderThanStr)
			if err != nil {
				log.Fatalf("invalid duration %q: %v", *olderThanStr, err)
			}
		}
		if err := cli.PruneRuns(db, olderThan, *command); err != nil {
			log.Fatalf("prune: %v", err)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n", args[0])
		os.Exit(1)
	}
}
