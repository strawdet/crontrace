package cli

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/crontrace/internal/store"
)

// ManageSchedule handles the `schedule` subcommand: set, del, list.
func ManageSchedule(db *sql.DB, args []string) error {
	repo, err := store.NewScheduleRepository(db)
	if err != nil {
		return fmt.Errorf("schedule: %w", err)
	}
	if len(args) == 0 {
		return scheduleList(repo)
	}
	switch args[0] {
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: schedule set <command> <cron_expr>")
		}
		return scheduleSet(repo, args[1], args[2])
	case "del":
		if len(args) < 2 {
			return fmt.Errorf("usage: schedule del <command>")
		}
		return scheduleDel(repo, args[1])
	case "list":
		return scheduleList(repo)
	default:
		return fmt.Errorf("schedule: unknown subcommand %q (set|del|list)", args[0])
	}
}

func scheduleSet(repo *store.ScheduleRepository, command, cronExpr string) error {
	if err := repo.Upsert(command, cronExpr); err != nil {
		return err
	}
	fmt.Printf("schedule set: %s → %s\n", command, cronExpr)
	return nil
}

func scheduleDel(repo *store.ScheduleRepository, command string) error {
	if err := repo.Delete(command); err != nil {
		return err
	}
	fmt.Printf("schedule deleted: %s\n", command)
	return nil
}

func scheduleList(repo *store.ScheduleRepository) error {
	schedules, err := repo.List()
	if err != nil {
		return err
	}
	if len(schedules) == 0 {
		fmt.Println("no schedules registered")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COMMAND\tCRON EXPR\tUPDATED")
	for _, s := range schedules {
		fmt.Fprintf(w, "%s\t%s\t%s\n", s.Command, s.CronExpr, s.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	return w.Flush()
}
