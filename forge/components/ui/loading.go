package ui

import (
	"encoding/json"
	"github.com/sojupie/tokens/forge/components"
)

// Render the runtime's source geometry before JavaScript starts animating it.
var initialLoadingPath = func() string {
	data, err := components.Files.ReadFile("static/loading-shapes.json")
	if err != nil {
		panic(err)
	}
	var shapes map[string]struct {
		Initial string `json:"initial"`
	}
	if err := json.Unmarshal(data, &shapes); err != nil {
		panic(err)
	}
	if shapes["indeterminate"].Initial == "" {
		panic("missing loading geometry")
	}
	return shapes["indeterminate"].Initial
}()
