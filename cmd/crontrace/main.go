// crontrace wraps cron job execution and records history, durations, and exit
// codes to a local SQLite database for later inspection.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yourorg/crontrace/internal/cli"
	"github.com/yourorg/crontrace/internal/runner"
	"github.com/yourorg/crontrace/internal/store"
)

// defaultDBPath returns the path to the default SQLite database file.
// It is placed in the user's home directory under .crontrace/jobs.db.
func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "crontrace.db"
	}
	return filepath.Join(home, ".crontrace", "jobs.db")
}

func main() {
	var dbPath string

	rootCmd := &cobra.Command{
		Use:   "crontrace",
		Short: "Record and inspect cron job execution history",
		Long: `crontrace wraps cron job commands, recording their exit codes and
durations to a local SQLite store so you can audit and debug scheduled tasks.`,
	}

	rootCmd.PersistentFlags().StringVar(
		&dbPath, "db", defaultDBPath(),
		"path to the SQLite database file",
	)

	// run subcommand — wraps a command and records its execution.
	runCmd := &cobra.Command{
		Use:                "run -- <command> [args...]",
		Short:              "Run a command and record its execution",
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			repo := store.NewJobRunRepository(db)
			exitCode, runErr := runner.Run(args[0], args[1:], repo)
			if runErr != nil {
				// Print the error but still exit with the recorded code.
				fmt.Fprintf(os.Stderr, "crontrace: %v\n", runErr)
			}
			os.Exit(exitCode)
			return nil
		},
	}

	// list subcommand — prints recent job run history.
	listCmd := &cobra.Command{
		Use:   "list [command-filter]",
		Short: "List recorded job runs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}
			return cli.ListRuns(db, os.Stdout, filter)
		},
	}

	// stats subcommand — prints aggregate statistics per command.
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show aggregate statistics for recorded commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := store.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			return cli.PrintStats(db, os.Stdout)
		},
	}

	rootCmd.AddCommand(runCmd, listCmd, statsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
