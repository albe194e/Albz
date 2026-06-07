package sqlc

import _ "embed"

// SchemaSQL embeds the runtime SQLite schema from the same file sqlc uses.
//
//go:embed schema.sql
var SchemaSQL string
