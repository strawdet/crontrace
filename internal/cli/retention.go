package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/user/crontrace/internal/store"
)

// ApplyRetention enforces a retention policy based on CLI args.
// Usage:
//
//	crontrace retention --max-age-days=N
//	crontrace retention --max-runs=N
func ApplyRetention(db *sql.DB, args []string) {
	if err := applyRetention(db, args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "retention error: %v\n", err)
		os.Exit(1)
	}
}

func applyRetention(db *sql.DB, args []string, out interface{ Write([]byte) (int, error) }) error {
	policy := store.RetentionPolicy{}

	for _, arg := range args {
		var key, val string
		for i, c := range arg {
			if c == '=' {
				key = arg[:i]
				val = arg[i+1:]
				break
			}
		}
		switch key {
		case "--max-age-days":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid --max-age-days value: %q", val)
			}
			policy.MaxAgeDays = n
		case "--max-runs":
			n, err := strconv.Atoi(val)
			if err != nil || n <= 0 {
				return fmt.Errorf("invalid --max-runs value: %q", val)
			}
			policy.MaxRunsPerCommand = n
		default:
			return fmt.Errorf("unknown retention flag: %q", arg)
		}
	}

	if policy.MaxAgeDays == 0 && policy.MaxRunsPerCommand == 0 {
		return fmt.Errorf("at least one of --max-age-days or --max-runs is required")
	}

	repo := store.NewRetentionRepository(db)
	deleted, err := repo.Apply(policy)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Retention applied: %d record(s) removed.\n", deleted)
	return nil
}
