package schemas

import "embed"

// Files contains the published Token Catalog JSON Schemas.
//
//go:embed *.json
var Files embed.FS
