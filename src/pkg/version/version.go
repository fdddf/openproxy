package version

import "fmt"

var (
	Commit    = "dev"
	BuildDate = "unknown"
)

// Info returns a human-readable version string.
func Info() string {
	return fmt.Sprintf("%s (built %s)", Commit, BuildDate)
}
