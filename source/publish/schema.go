package publish

import _ "embed"

// SchemaSQL is the generated publication database schema.
//
//go:embed schema.sql
var SchemaSQL string
