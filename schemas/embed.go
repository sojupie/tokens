package schemas

import "embed"

// Files contains the published Source Loom JSON Schemas.
//
//go:embed *.json
var Files embed.FS
