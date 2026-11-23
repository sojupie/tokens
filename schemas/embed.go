package schemas

import "embed"

// Files contains the published TokenCatalog JSON Schemas.
//
//go:embed *.json
var Files embed.FS
