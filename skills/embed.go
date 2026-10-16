// Package skills embeds the guides served by the catalog.
package skills

import _ "embed"

// MaterialGuide is shared by the hosted agent guide and installable skill.
//
//go:embed token-catalog/references/material-tokens.md
var MaterialGuide string
