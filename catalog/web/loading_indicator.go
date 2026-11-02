package web

import (
	"github.com/sojupie/tokens/forge/components"
)

// Catalog versions the geometry asset; Forge owns its interpretation and first frame.
var loadingGeometryVersion = func() string {
	data, err := components.Files.ReadFile("static/loading-shapes.json")
	if err != nil {
		panic(err)
	}
	return assetVersion(data)
}()
