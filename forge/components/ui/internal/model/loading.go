package model

import (
	"encoding/json"
	"github.com/sojupie/tokens/forge/components"
)

// Render the runtime's source geometry before JavaScript starts animating it.
//
//forge:internal
var LoadingPaths = func() map[string]string {
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
	return map[string]string{"determinate": shapes["determinate"].Initial, "indeterminate": shapes["indeterminate"].Initial}
}()

//forge:internal
var InitialLoadingPath = LoadingPaths["indeterminate"]

//forge:internal
func LoadingRotation(control string) string {
	if control == "determinate" {
		return "rotate(-72 .5 .5)"
	}
	return "rotate(90 .5 .5)"
}
