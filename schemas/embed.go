package schemas

import "embed"

// Files contains the published Tokens JSON Schemas.
//
//go:embed *.json
var Files embed.FS
