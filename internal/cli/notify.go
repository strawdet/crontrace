package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/user/crontrace/internal/store"
)

// ManageNotify is the entry point for the `crontrace notify` subcommand.
// args: [set <command> --max-ms <ms> --alert-on-fail] | [del <command>] | [list]
func ManageNotify(db *sql.DB, args []string) error {
	repo, err := store.NewNotifyRepository(db)
	if err != nil {
		return fmt.Errorf("notify: open repository: %w", err)
	}
	if len(args) == 0 {
		return fmt.Errorf("notify: expected subcommand: set, del, list")
	}
	switch args[0] {
	case "set":
		return notifySet(repo, args[1:])
	case "del":
		return notifyDel(repo, args[1:])
	case "list":
		return notifyList(repo)
	default:
		return fmt.Errorf("notify: unknown subcommand %q", args[0])
	}
}

func notifySet(repo *store.NotifyRepository, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("notify set: expected <command>")
	}
	rule := store.NotifyRule{
		Command:     args[0],
		AlertOnFail: true,
	}
	for i := 1; i < len(args)-1; i++ {
		switch args[i] {
		case "--max-ms":
			ms, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return fmt.Errorf("notify set: invalid --max-ms value: %w", err)
			}
			rule.MaxDurationMs = ms
			i++
		case "--no-alert-on-fail":
			rule.AlertOnFail = false
		}
	}
	if err := repo.Upsert(rule); err != nil {
		return fmt.Errorf("notify set: %w", err)
	}
	fmt.Printf("Notification rule set for %q (max-ms=%d, alert-on-fail=%v)\n",
		rule.Command, rule.MaxDurationMs, rule.AlertOnFail)
	return nil
}

func notifyDel(repo *store.NotifyRepository, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("notify del: expected <command>")
	}
	if err := repo.Delete(args[0]); err != nil {
		return fmt.Errorf("notify del: %w", err)
	}
	fmt.Printf("Notification rule removed for %q\n", args[0])
	return nil
}

func notifyList(repo *store.NotifyRepository) error {
	rules, err := repo.List()
	if err != nil {
		return fmt.Errorf("notify list: %w", err)
	}
	if len(rules) == 0 {
		fmt.Println("No notification rules configured.")
		return nil
	}
	fmt.Printf("%-40s %12s %14s\n", "COMMAND", "MAX-MS", "ALERT-ON-FAIL")
	for _, r := range rules {
		fmt.Printf("%-40s %12d %14v\n", r.Command, r.MaxDurationMs, r.AlertOnFail)
	}
	return nil
}
