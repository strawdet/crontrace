package cli

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/user/crontrace/internal/store"
)

var validMetrics = map[string]bool{
	"duration":     true,
	"exit_code":    true,
	"failure_rate": true,
}

// ManageAlerts dispatches alert subcommands: set, del, list.
func ManageAlerts(db *sql.DB, args []string) error {
	repo, err := store.NewAlertRepository(db)
	if err != nil {
		return fmt.Errorf("alert repo: %w", err)
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: alert <set|del|list> [args...]")
	}
	switch args[0] {
	case "set":
		return alertSet(repo, args[1:])
	case "del":
		return alertDel(repo, args[1:])
	case "list":
		return alertList(repo, args[1:])
	default:
		return fmt.Errorf("unknown alert subcommand: %s", args[0])
	}
}

func alertSet(repo *store.AlertRepository, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: alert set <command> <metric> <threshold>")
	}
	command, metric := args[0], args[1]
	if !validMetrics[metric] {
		return fmt.Errorf("invalid metric %q; choose from: duration, exit_code, failure_rate", metric)
	}
	threshold, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return fmt.Errorf("invalid threshold %q: %w", args[2], err)
	}
	if err := repo.Upsert(command, metric, threshold); err != nil {
		return fmt.Errorf("set alert: %w", err)
	}
	fmt.Printf("Alert set: %s %s >= %.4g\n", command, metric, threshold)
	return nil
}

func alertDel(repo *store.AlertRepository, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: alert del <command> <metric>")
	}
	n, err := repo.Delete(args[0], args[1])
	if err != nil {
		return fmt.Errorf("del alert: %w", err)
	}
	if n == 0 {
		fmt.Println("No matching alert found.")
	} else {
		fmt.Printf("Deleted alert: %s %s\n", args[0], args[1])
	}
	return nil
}

func alertList(repo *store.AlertRepository, args []string) error {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}
	alerts, err := repo.List(command)
	if err != nil {
		return fmt.Errorf("list alerts: %w", err)
	}
	if len(alerts) == 0 {
		fmt.Println("No alerts configured.")
		return nil
	}
	fmt.Printf("%-30s %-14s %s\n", "COMMAND", "METRIC", "THRESHOLD")
	for _, a := range alerts {
		fmt.Printf("%-30s %-14s %.4g\n", a.Command, a.Metric, a.Threshold)
	}
	return nil
}
