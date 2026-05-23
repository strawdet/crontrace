package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/user/crontrace/internal/store"
)

// ManageWatch handles the `watch` subcommand: set, del, list.
func ManageWatch(db *sql.DB, args []string) error {
	repo, err := store.NewWatchRepository(db)
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}
	if len(args) == 0 {
		return manageWatchList(repo)
	}
	switch args[0] {
	case "set":
		return manageWatchSet(repo, args[1:])
	case "del":
		return manageWatchDel(repo, args[1:])
	case "list":
		return manageWatchList(repo)
	default:
		return fmt.Errorf("watch: unknown subcommand %q (use set, del, list)", args[0])
	}
}

func manageWatchSet(repo *store.WatchRepository, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("watch set: usage: watch set <command> [max_duration_sec] [expect_exit_code]")
	}
	cmd := args[0]
	maxDur := 0
	expectExit := 0
	var err error
	if len(args) >= 2 {
		maxDur, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("watch set: invalid max_duration_sec: %w", err)
		}
	}
	if len(args) >= 3 {
		expectExit, err = strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("watch set: invalid expect_exit_code: %w", err)
		}
	}
	if err := repo.Upsert(cmd, maxDur, expectExit); err != nil {
		return fmt.Errorf("watch set: %w", err)
	}
	fmt.Printf("watch: set %q (max_duration=%ds, expect_exit=%d)\n", cmd, maxDur, expectExit)
	return nil
}

func manageWatchDel(repo *store.WatchRepository, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("watch del: usage: watch del <command>")
	}
	if err := repo.Delete(args[0]); err != nil {
		return fmt.Errorf("watch del: %w", err)
	}
	fmt.Printf("watch: removed %q\n", args[0])
	return nil
}

func manageWatchList(repo *store.WatchRepository) error {
	entries, err := repo.List()
	if err != nil {
		return fmt.Errorf("watch list: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No watches configured.")
		return nil
	}
	fmt.Printf("%-40s %12s %16s\n", "COMMAND", "MAX_DUR(s)", "EXPECT_EXIT")
	for _, e := range entries {
		fmt.Printf("%-40s %12d %16d\n", e.Command, e.MaxDurationSec, e.ExpectExitCode)
	}
	return nil
}
