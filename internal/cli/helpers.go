package cli

import "time"

// nowUTC returns the current UTC time, extracted so tests can stub it if needed.
func nowUTC() time.Time {
	return time.Now().UTC()
}
