// Package migrations embeds the SQL migration sets so the binary can run
// migrations without the files being present on disk next to it.
package migrations

import "embed"

//go:embed postgres/*.sql sqlite/*.sql
var FS embed.FS
