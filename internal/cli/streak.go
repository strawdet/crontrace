package cli

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/user/crontrace/internal/store"
)

// PrintStreak prints the current run streak for a given command.
func PrintStreak(db *sql.DB, command string) error {
	return printStreak(db, command, os.Stdout)
}

func printStreak(db *sql.DB, command string, w io.Writer) error {
	if command == "" {
		return fmt.Errorf("streak: command argument is required")
	}

	repo := store.NewStreakRepository(db)
	result, err := repo.GetStreak(command)
	if err != nil {
		return fmt.Errorf("streak: query failed: %w", err)
	}

	if result == nil {
		fmt.Fprintf(w, "No runs found for command: %s\n", command)
		return nil
	}

	icon := "✅"
	if result.StreakType == "failure" {
		icon = "❌"
	}

	fmt.Fprintf(w, "Command : %s\n", result.Command)
	fmt.Fprintf(w, "Streak  : %s %d consecutive %s(s)\n", icon, result.CurrentStreak, result.StreakType)
	fmt.Fprintf(w, "Last Run: %s\n", result.LastRun.Format("2006-01-02 15:04:05 UTC"))
	return nil
}
