package runner

import (
	"os/exec"
	"time"

	"github.com/user/crontrace/internal/store"
)

// Result holds the outcome of a single job execution.
type Result struct {
	Command  string
	Args     []string
	Started  time.Time
	Finished time.Time
	ExitCode int
	Output   []byte
}

// Run executes the given command with its arguments, records timing and exit
// code, and persists the result via the provided repository.
func Run(repo *store.JobRunRepository, command string, args []string) (*Result, error) {
	cmd := exec.Command(command, args...)

	started := time.Now()
	output, err := cmd.CombinedOutput()
	finished := time.Now()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Command could not be started at all.
			return nil, err
		}
	}

	result := &Result{
		Command:  command,
		Args:     args,
		Started:  started,
		Finished: finished,
		ExitCode: exitCode,
		Output:   output,
	}

	run := store.JobRun{
		Command:    command,
		StartedAt:  started,
		FinishedAt: finished,
		ExitCode:   exitCode,
		Output:     string(output),
	}

	if insertErr := repo.Insert(run); insertErr != nil {
		return result, insertErr
	}

	return result, nil
}
