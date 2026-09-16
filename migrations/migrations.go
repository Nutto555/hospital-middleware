// Package migrations embeds the versioned SQL applied at service start.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
