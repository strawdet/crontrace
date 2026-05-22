package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/user/crontrace/internal/store"
)

// ManageTags handles the `crontrace tag` subcommand for adding, removing, and listing tags.
func ManageTags(db *sql.DB, action string, runID int64, tags []string) error {
	repo, err := store.NewTagRepository(db)
	if err != nil {
		return fmt.Errorf("init tag repository: %w", err)
	}
	return manageTags(repo, action, runID, tags)
}

type tagRepository interface {
	AddTag(runID int64, tag string) error
	RemoveTag(runID int64, tag string) error
	ListTags(runID int64) ([]string, error)
	RunIDsByTag(tag string) ([]int64, error)
}

func manageTags(repo tagRepository, action string, runID int64, tags []string) error {
	switch strings.ToLower(action) {
	case "add":
		if len(tags) == 0 {
			return fmt.Errorf("add requires at least one tag")
		}
		for _, tag := range tags {
			if err := repo.AddTag(runID, tag); err != nil {
				return fmt.Errorf("add tag %q to run %d: %w", tag, runID, err)
			}
			fmt.Printf("Tagged run %d with %q\n", runID, tag)
		}

	case "remove":
		if len(tags) == 0 {
			return fmt.Errorf("remove requires at least one tag")
		}
		for _, tag := range tags {
			if err := repo.RemoveTag(runID, tag); err != nil {
				return fmt.Errorf("remove tag %q from run %d: %w", tag, runID, err)
			}
			fmt.Printf("Removed tag %q from run %d\n", runID, tag)
		}

	case "list":
		result, err := repo.ListTags(runID)
		if err != nil {
			return fmt.Errorf("list tags for run %d: %w", runID, err)
		}
		if len(result) == 0 {
			fmt.Printf("No tags for run %d\n", runID)
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TAG")
		for _, t := range result {
			fmt.Fprintln(w, t)
		}
		w.Flush()

	default:
		return fmt.Errorf("unknown tag action %q: use add, remove, or list", action)
	}
	return nil
}
