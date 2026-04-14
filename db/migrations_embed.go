package dbfiles

import "embed"

// FS exposes SQL assets for runtime application startup and tests.
//
//go:embed migrations/*.sql seeds/*.sql
var FS embed.FS
