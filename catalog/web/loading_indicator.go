package web

import (
	"encoding/json"
	"io/fs"

	"github.com/sojupie/tokens/forge/components"
)

// Keep the first frame server-rendered, including inert catalog illustrations.
var loadingGeometry = func() struct {
	initial map[string]string
	version string
} { data, err := fs.ReadFile(components.Files, "static/loading-shapes.json"); if err != nil {
	panic(err)
}; var sequences map[string]struct {
	Initial string `json:"initial"`
}; if err = json.Unmarshal(data, &sequences); err != nil {
	panic(err)
}; result := struct {
	initial map[string]string
	version string
}{map[string]string{}, assetVersion(data)}; for _, control := range []string{"determinate", "indeterminate"} {
	if sequences[control].Initial == "" {
		panic("missing loading indicator geometry")
	}
	result.initial[control] = sequences[control].Initial
}; return result }()

func loadingInitialPath(control string) string { return loadingGeometry.initial[control] }
func loadingRotation(control string) string {
	if control == "determinate" {
		return "rotate(-72 .5 .5)"
	}
	return "rotate(90 .5 .5)"
}
